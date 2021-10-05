package rest

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func getOffset(query url.Values) (uint64, error) {
	return getUint(query, "pagination.offset")
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
		if !strings.Contains(v, "=") {
			return nil, fmt.Errorf("query event missing equal sign: %s", v)
		}
			arr := strings.SplitN(v, "=", 2)
			value, err := trimSingleQuotes(arr[1])
			if err != nil {
				return nil, err
			}
		key := arr[0]
		if m[key] != nil {
			return nil, fmt.Errorf("event appears more than once: %s", key)
		}
		m[key] = []string{value}
	}
	return m, nil
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
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	defer conn.Release()
	totalCount, err := db.QueryCount(conn, events)
	if err != nil {
		logger.L.Errorw("Cannot get total tx count from database", "events", events, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	txResponses, err := db.QueryTxs(conn, events, limit, offsetInTimesOfLimit, orderByDesc)
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
