package rest

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/util"
)

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func getEvents(query url.Values) (events []util.Event) {
	for k, vs := range query {
		if strings.Contains(k, ".") {
			arr := strings.SplitN(k, ".", 2)
			for _, v := range vs {
				events = append(events, util.Event{
					Type: arr[0],
					Attributes: []util.Attribute{
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

func queryCount(conn *pgxpool.Conn, events []util.Event) (int64, error) {
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

func queryTxs(conn *pgxpool.Conn, events []util.Event, limit int64, page int64) ([]interface{}, error) {
	offset := limit * (page - 1)
	sql := `
		SELECT tx FROM txs
		WHERE event_hashes @> $1
		ORDER BY height, tx_index
		LIMIT $2
		OFFSET $3
	`
	eventHashes := util.GetEventHashes(events)
	ctx, cancel := db.GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, eventHashes, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]interface{}, 0, limit)
	for rows.Next() {
		txRow, err := rows.Values()
		if err != nil {
			return nil, err
		}
		res = append(res, txRow[0])
	}
	return res, nil
}

type Response struct {
	TotalCount string        `json:"total_count"`
	Count      string        `json:"count"`
	Page       string        `json:"page_number"`
	PageTotal  string        `json:"page_total"`
	Limit      string        `json:"limit"`
	Txs        []interface{} `json:"txs"`
}

func handleTxsSearch(c *gin.Context, pool *pgxpool.Pool) {
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
	txs, err := queryTxs(conn, events, limit, page)
	if err != nil {
		logger.L.Errorw("Cannot get txs from database", "events", events, "limit", limit, "page", page, "error", err)
		c.AbortWithStatusJSON(500, err)
		return
	}
	c.JSON(200, Response{
		TotalCount: fmt.Sprintf("%d", totalCount),
		Count:      fmt.Sprintf("%d", len(txs)),
		Page:       fmt.Sprintf("%d", page),
		PageTotal:  fmt.Sprintf("%d", totalPages),
		Limit:      fmt.Sprintf("%d", limit),
		Txs:        txs,
	})
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
			handleTxsSearch(c, pool)
			return
		}
		proxyHandler(c)
	})
	router.POST("/*endpoint", proxyHandler)
	router.Run(listenAddr)
}
