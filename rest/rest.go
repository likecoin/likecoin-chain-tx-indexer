package rest

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

type Attribute struct {
	Type  string
	Key   string
	Value string
}

func getAttributes(query url.Values) (attrs []Attribute) {
	for k, vs := range query {
		if strings.Contains(k, ".") {
			arr := strings.SplitN(k, ".", 2)
			for _, v := range vs {
				attrs = append(attrs, Attribute{
					Type:  arr[0],
					Key:   arr[1],
					Value: v,
				})
			}
		}
	}
	return attrs
}

func queryCount(conn *pgx.Conn, attr Attribute) (int64, error) {
	sql := `
SELECT count(DISTINCT (height, tx_index)) FROM tx_events
WHERE type = $1 AND key = $2 AND value = $3
`
	row := conn.QueryRow(context.Background(), sql, attr.Type, attr.Key, attr.Value)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func queryTxs(conn *pgx.Conn, attr Attribute, limit int64, page int64) ([]interface{}, error) {
	offset := limit * (page - 1)
	sql := `
SELECT txs.tx FROM (
	SELECT DISTINCT height, tx_index FROM tx_events
	WHERE type = $1 AND key = $2 AND value = $3
	ORDER BY height, tx_index
	LIMIT $4
	OFFSET $5
) as e
JOIN txs ON txs.height = e.height AND txs.tx_index = e.tx_index
`
	rows, err := conn.Query(context.Background(), sql, attr.Type, attr.Key, attr.Value, limit, offset)
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

func Run(conn *pgx.Conn, listenAddr string) {
	router := gin.New()
	router.GET("/txs", func(c *gin.Context) {
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
		attrs := getAttributes(q)
		if len(attrs) == 0 {
			c.AbortWithStatusJSON(400, "Need attribute")
			return
		}
		if len(attrs) > 1 {
			c.AbortWithStatusJSON(400, "only 1 attribute supported")
			return
		}
		attr := attrs[0]
		totalCount, err := queryCount(conn, attr)
		if err != nil {
			c.AbortWithError(500, err)
		}
		totalPages := (totalCount-1)/limit + 1
		txs, err := queryTxs(conn, attr, limit, page)
		if err != nil {
			c.AbortWithError(500, err)
		}
		c.JSON(200, Response{
			TotalCount: fmt.Sprintf("%d", totalCount),
			Count:      fmt.Sprintf("%d", len(txs)),
			Page:       fmt.Sprintf("%d", page),
			PageTotal:  fmt.Sprintf("%d", totalPages),
			Limit:      fmt.Sprintf("%d", limit),
			Txs:        txs,
		})
	})
	router.Run(listenAddr)
}
