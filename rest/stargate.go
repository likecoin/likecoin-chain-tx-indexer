package rest

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func getEventMapAndHeight(eventArray []string) (url.Values, uint64, error) {
	var height = uint64(0)
	var err error
	m := make(url.Values)
	for _, v := range eventArray {
		if !strings.Contains(v, "=") {
			return nil, 0, fmt.Errorf("query event missing equal sign: %s", v)
		}
		arr := strings.SplitN(v, "=", 2)
		if arr[0] == "tx.height" {
			height, err = strconv.ParseUint(arr[1], 10, 64)
			if err != nil {
				return nil, 0, err
			}
		} else {
			value, err := trimSingleQuotes(arr[1])
			if err != nil {
				return nil, 0, err
			}
			key := arr[0]
			if m[key] != nil {
				return nil, 0, fmt.Errorf("event appears more than once: %s", key)
			}
			m[key] = []string{value}
		}
	}
	return m, height, nil
}

func handleStargateTxsSearch(c *gin.Context) {
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
	reverse := getReverse(q)

	key, err := getKey(q)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	shouldCountTotal, err := getBool(q, "pagination.count_total")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	eventArray := c.QueryArray("events")
	eventMap, height, err := getEventMapAndHeight(eventArray)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	events, err := getEvents(eventMap)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn := getConn(c)

	var totalCount uint64
	if shouldCountTotal {
		totalCount, err = db.QueryCount(conn, events, height)
		if err != nil {
			logger.L.Errorw("Cannot get total tx count from database", "events", events, "error", err)
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	p := db.PageRequest{
		Key:     key,
		Limit:   int(limit),
		Reverse: reverse,
		Offset:  offsetInTimesOfLimit,
	}
	nextKey, txResponses, err := db.QueryTxs(conn, events, height, p)
	if err != nil {
		logger.L.Errorw("Cannot get txs from database", "events", events, "limit", limit, "offset", offset, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	pagination := &query.PageResponse{
		Total:   totalCount,
		NextKey: nextKey,
	}
	res := tx.GetTxsEventResponse{
		TxResponses: txResponses,
		Pagination:  pagination,
	}

	for _, txResponse := range txResponses {
		var tx tx.Tx
		err := tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			logger.L.Warn("Cannot unmarshal tx response", "tx", txResponse.Tx.Value, "error", err)
			continue
		}
		res.Txs = append(res.Txs, &tx)
	}

	resJson, err := encodingConfig.Marshaler.MarshalJSON(&res)
	if err != nil {
		logger.L.Errorw("Cannot marshal GetTxsEventResponse to JSON", "events", events, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	_, _ = c.Writer.Write(resJson)
}
