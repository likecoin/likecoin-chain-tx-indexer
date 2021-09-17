package rest

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likechain/app"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/util"
)

var encodingConfig = app.MakeEncodingConfig()

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func getEvents(query url.Values) (events types.StringEvents) {
	for k, vs := range query {
		if strings.Contains(k, ".") {
			arr := strings.SplitN(k, ".", 2)
			for _, v := range vs {
				events = append(events, types.StringEvent{
					Type: arr[0],
					Attributes: []types.Attribute{
						{
							Key:   arr[1],
							Value: trimQuotes(v),
						},
					},
				})
			}
		}
	}
	return events
}

func queryCount(conn *pgxpool.Conn, events types.StringEvents) (int64, error) {
	sql := `
		SELECT count(id) FROM txs
		WHERE event_hashes @> $1
	`
	ctx, cancel := db.GetTimeoutContext()
	defer cancel()
	eventHashes := util.GetEventHashes(events)
	row := conn.QueryRow(ctx, sql, eventHashes)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func queryTxs(conn *pgxpool.Conn, events types.StringEvents, limit int64, page int64, orderByDesc bool) ([]*types.TxResponse, error) {
	offset := limit * (page - 1)
	order := "ASC"
	if orderByDesc {
		order = "DESC"
	}
	sql := fmt.Sprintf(`
		SELECT tx FROM txs
		WHERE event_hashes @> $1
		ORDER BY height %s, tx_index %s
		LIMIT $2
		OFFSET $3
	`, order, order)
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
		txRow, err := rows.Values()
		if err != nil {
			return nil, err
		}
		txJson, ok := txRow[0].(*pgtype.JSONB)
		if !ok {
			return nil, fmt.Errorf("expecting a type implementing intoAny, got: %T", txRow[0])
		}
		if txJson.Status != pgtype.Present {
			return nil, fmt.Errorf("cannot unmarshal JSON to TxResponse: %+v", txJson.Status)
		}
		var txRes types.TxResponse
		err = encodingConfig.Marshaler.UnmarshalJSON(txJson.Bytes, &txRes)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal JSON to TxResponse: %+v", txJson)
		}
		res = append(res, &txRes)
	}
	return res, nil
}

type AminoResponse struct {
	TotalCount string        `json:"total_count"`
	Count      string        `json:"count"`
	Page       string        `json:"page_number"`
	PageTotal  string        `json:"page_total"`
	Limit      string        `json:"limit"`
	Txs        []interface{} `json:"txs"`
}

func handleAminoTxsSearch(c *gin.Context, pool *pgxpool.Pool) {
	q := c.Request.URL.Query()
	var page int64
	var limit int64
	var err error
	pageStr := q.Get("page")
	if pageStr == "" {
		page = 1
	} else {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil || page < 1 {
			c.AbortWithStatus(400)
			return
		}
	}
	limitStr := q.Get("limit")
	if limitStr == "" {
		limit = 1
	} else {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit < 1 || limit > 100 {
			c.AbortWithStatus(400)
			return
		}
	}
	events := getEvents(q)
	if len(events) == 0 {
		c.AbortWithStatusJSON(400, "event needed")
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
		c.AbortWithStatusJSON(500, err)
		return
	}
	totalPages := (totalCount-1)/limit + 1
	txs, err := queryTxs(conn, events, limit, page, false)
	if err != nil {
		logger.L.Errorw("Cannot get txs from database", "events", events, "limit", limit, "page", page, "error", err)
		c.AbortWithStatusJSON(500, err)
		return
	}
	c.JSON(200, AminoResponse{
		TotalCount: fmt.Sprintf("%d", totalCount),
		Count:      fmt.Sprintf("%d", len(txs)),
		Page:       fmt.Sprintf("%d", page),
		PageTotal:  fmt.Sprintf("%d", totalPages),
		Limit:      fmt.Sprintf("%d", limit),
		Txs:        txs,
	})
}

func getEventMap(eventArray []string) (map[string][]string, error) {
	m := make(map[string][]string)
	for _, v := range eventArray {
		if strings.Contains(v, "=") {
			arr := strings.SplitN(v, "=", 2)
			m[arr[0]] = []string{arr[1]}
		} else {
			return nil, fmt.Errorf("query event missing equal sign: %s", v)
		}
	}
	return m, nil
}

func handleStargateTxsSearch(c *gin.Context, pool *pgxpool.Pool) {
	q := c.Request.URL.Query()
	var offset int64
	var limit int64
	var err error
	orderByDesc := false
	offsetStr := q.Get("pagination.offset")
	if offsetStr == "" {
		offset = 1
	} else {
		offset, err = strconv.ParseInt(offsetStr, 10, 64)
		if err != nil || offset < 1 {
			c.AbortWithStatus(400)
			return
		}
	}
	limitStr := q.Get("pagination.limit")
	if limitStr == "" {
		limit = 1
	} else {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit < 1 || limit > 100 {
			c.AbortWithStatus(400)
			return
		}
	}
	orderByStr := strings.ToUpper(q.Get("order_by"))
	switch orderByStr {
	case "", "ORDER_BY_UNSPECIFIED", "ORDER_BY_ASC":
		break
	case "ORDER_BY_DESC":
		orderByDesc = true
	default:
		c.AbortWithStatusJSON(400, "Available values for order_by: ORDER_BY_UNSPECIFIED, ORDER_BY_ASC, ORDER_BY_DESC")
		return
	}
	eventArray := c.QueryArray("events")
	if len(eventArray) == 0 {
		c.AbortWithStatusJSON(400, "event needed")
		return
	}
	eventMap, err := getEventMap(eventArray)
	if err != nil {
		c.AbortWithStatusJSON(400, err.Error())
		return
	}
	events := getEvents(eventMap)

	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		logger.L.Panicw("Cannot acquire connection from database connection pool", "error", err)
	}
	defer conn.Release()
	totalCount, err := queryCount(conn, events)
	if err != nil {
		logger.L.Errorw("Cannot get total tx count from database", "events", events, "error", err)
		c.AbortWithStatusJSON(500, err)
		return
	}
	txResponses, err := queryTxs(conn, events, limit, offset, orderByDesc)
	if err != nil {
		logger.L.Errorw("Cannot get txs from database", "events", events, "limit", limit, "page", offset, "error", err)
		c.AbortWithStatusJSON(500, err)
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

	resJson, _ := encodingConfig.Marshaler.MarshalJSON(&res)
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
