package rest

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likechain/app"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/util"
)

var encodingConfig = app.MakeEncodingConfig()

func getUint(query url.Values, key string) (uint64, error) {
	valueStr := query.Get(key)
	if valueStr == "" {
		return 0, nil
	} else {
		return strconv.ParseUint(valueStr, 10, 64)
	}
}

func getPage(query url.Values) (uint64, error) {
	page, err := getUint(query, "page")
	if page == 0 {
		page = 1
	}
	return page, err
}

func getOffset(query url.Values) (uint64, error) {
	return getUint(query, "pagination.offset")
}

func getLimit(query url.Values, key string) (uint64, error) {
	limit, err := getUint(query, key)
	if limit == 0 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	return limit, err
}

func isOrderByDesc(query url.Values) (bool, error) {
	orderByStr := strings.ToUpper(query.Get("order_by"))
	switch orderByStr {
	case "", "ORDER_BY_UNSPECIFIED", "ORDER_BY_ASC":
		return false, nil
	case "ORDER_BY_DESC":
		return true, nil
	default:
		return false, fmt.Errorf("available values for order_by: ORDER_BY_UNSPECIFIED, ORDER_BY_ASC, ORDER_BY_DESC")
	}
}

func trimSingleQuotes(s string) (string, error) {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && c == '\'' {
			return s[1 : len(s)-1], nil
		}
		return "", fmt.Errorf("expect query event value missing single quotes: %s", s)
	}
	return "", fmt.Errorf("invalid query event value: %s", s)
}

func getEventMap(eventArray []string) (url.Values, error) {
	m := make(url.Values)
	for _, v := range eventArray {
		if strings.Contains(v, "=") {
			arr := strings.SplitN(v, "=", 2)
			value, err := trimSingleQuotes(arr[1])
			if err != nil {
				return nil, err
			}
			m[arr[0]] = []string{value}
		} else {
			return nil, fmt.Errorf("query event missing equal sign: %s", v)
		}
	}
	return m, nil
}

func getEvents(query url.Values) (events types.StringEvents, err error) {
	for k, vs := range query {
		if strings.Contains(k, ".") {
			arr := strings.SplitN(k, ".", 2)
			for _, v := range vs {
				events = append(events, types.StringEvent{
					Type: arr[0],
					Attributes: []types.Attribute{
						{
							Key:   arr[1],
							Value: v,
						},
					},
				})
			}
		}
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("events needed")
	}
	return events, nil
}

func queryCount(conn *pgxpool.Conn, events types.StringEvents) (uint64, error) {
	sql := `
		SELECT count(*) FROM txs
		WHERE event_hashes @> $1
	`
	ctx, cancel := db.GetTimeoutContext()
	defer cancel()
	eventHashes := util.GetEventHashes(events)
	row := conn.QueryRow(ctx, sql, eventHashes)
	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func queryTxs(conn *pgxpool.Conn, events types.StringEvents, limit uint64, offset uint64, orderByDesc bool) ([]*types.TxResponse, error) {
	order := "ASC"
	if orderByDesc {
		order = "DESC"
	}
	sql := fmt.Sprintf(`
		SELECT tx FROM txs
		WHERE event_hashes @> $1
		ORDER BY id %s
		LIMIT $2
		OFFSET $3
	`, order)
	eventHashes := util.GetEventHashes(events)
	ctx, cancel := db.GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, eventHashes, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]*types.TxResponse, 0, limit)
	for rows.Next() {
		var jsonb pgtype.JSONB
		err := rows.Scan(&jsonb)
		if err != nil {
			return nil, err
		}
		var txRes types.TxResponse
		err = encodingConfig.Marshaler.UnmarshalJSON(jsonb.Bytes, &txRes)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal JSON to TxResponse: %+v", jsonb.Bytes)
		}
		res = append(res, &txRes)
	}
	return res, nil
}

func convertToStdTx(txBytes []byte) (legacytx.StdTx, error) {
	txI, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return legacytx.StdTx{}, err
	}

	tx, ok := txI.(signing.Tx)
	if !ok {
		return legacytx.StdTx{}, fmt.Errorf("%+v is not backwards compatible with %T", tx, legacytx.StdTx{})
	}

	return clienttx.ConvertTxToStdTx(encodingConfig.Amino, tx)
}

func packStdTxResponse(txRes *types.TxResponse) error {
	txBytes := txRes.Tx.Value
	stdTx, err := convertToStdTx(txBytes)
	if err != nil {
		return err
	}

	// Pack the amino stdTx into the TxResponse's Any.
	txRes.Tx = codectypes.UnsafePackAny(stdTx)
	return nil
}

func handleAminoTxsSearch(c *gin.Context, pool *pgxpool.Pool) {
	q := c.Request.URL.Query()
	page, err := getPage(q)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	limit, err := getLimit(q, "limit")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	events, err := getEvents(q)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		logger.L.Panicw("Cannot acquire connection from database connection pool", "error", err)
	}
	defer conn.Release()
	totalCount, err := queryCount(conn, events)
	if err != nil {
		logger.L.Errorw("Cannot get total tx count from database", "events", events, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	maxPage := (totalCount-1)/limit + 1
	if maxPage == 0 {
		maxPage = 1
	}
	if page > maxPage {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("page should be within [1, %d] range, given %d", maxPage, page)})
		return
	}
	offset := limit * (page - 1)
	txs, err := queryTxs(conn, events, limit, offset, false)
	if err != nil {
		logger.L.Errorw("Cannot get txs from database", "events", events, "limit", limit, "page", page, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	searchTxsResult := types.NewSearchTxsResult(totalCount, uint64(len(txs)), page, limit, txs)

	for _, txRes := range searchTxsResult.Txs {
		packStdTxResponse(txRes)
	}

	json, err := encodingConfig.Amino.MarshalJSON(searchTxsResult)
	if err != nil {
		logger.L.Errorw("Cannot convert searchTxsResult to JSON", "events", events, "limit", limit, "page", page, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(json)
}

func handleStargateTxsSearch(c *gin.Context, pool *pgxpool.Pool) {
	q := c.Request.URL.Query()

	offset, err := getOffset(q)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	limit, err := getLimit(q, "pagination.limit")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	offsetInTimesOfLimit := offset / limit * limit // Cosmos' bug? 29 / 10 * 10 = 20
	orderByDesc, err := isOrderByDesc(q)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	eventArray := c.QueryArray("events")
	eventMap, err := getEventMap(eventArray)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	events, err := getEvents(eventMap)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		logger.L.Panicw("Cannot acquire connection from database connection pool", "error", err)
	}
	defer conn.Release()
	totalCount, err := queryCount(conn, events)
	if err != nil {
		logger.L.Errorw("Cannot get total tx count from database", "events", events, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	txResponses, err := queryTxs(conn, events, limit, offsetInTimesOfLimit, orderByDesc)
	if err != nil {
		logger.L.Errorw("Cannot get txs from database", "events", events, "limit", limit, "offset", offset, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	res := tx.GetTxsEventResponse{
		TxResponses: txResponses,
		Pagination: &query.PageResponse{
			Total: uint64(totalCount),
		},
	}

	for _, txResponse := range txResponses {
		var tx tx.Tx
		tx.Unmarshal(txResponse.Tx.Value)
		res.Txs = append(res.Txs, &tx)
	}

	resJson, err := encodingConfig.Marshaler.MarshalJSON(&res)
	if err != nil {
		logger.L.Errorw("Cannot marshal GetTxsEventResponse to JSON", "events", events, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(resJson)
}

func Run(pool *pgxpool.Pool, listenAddr string, lcdEndpoint string) {
	lcdURL, err := url.Parse(lcdEndpoint)
	if err != nil {
		logger.L.Panicw("Cannot parse lcd URL", "lcd_endpoint", lcdEndpoint, "error", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(lcdURL)
	proxyHandler := func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	router := gin.New()
	router.GET("/*endpoint", func(c *gin.Context) {
		endpoint, ok := c.Params.Get("endpoint")
		if !ok {
			// ??? Gin bug?
			c.AbortWithStatus(500)
			return
		}
		if endpoint == "/txs" {
			handleAminoTxsSearch(c, pool)
			return
		}
		if endpoint == "/cosmos/tx/v1beta1/txs" {
			handleStargateTxsSearch(c, pool)
			return
		}
		proxyHandler(c)
	})
	router.POST("/*endpoint", proxyHandler)
	router.Run(listenAddr)
}
