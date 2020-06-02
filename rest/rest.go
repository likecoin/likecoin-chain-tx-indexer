package rest

import (
	"context"
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
		offset := limit * (page - 1)
		attrs := getAttributes(q)
		if len(attrs) == 0 {
			c.AbortWithStatus(400)
			return
		}
		if len(attrs) > 1 {
			c.AbortWithStatusJSON(400, "only 1 attribute supported")
			return
		}
		attr := attrs[0]
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
			c.AbortWithError(500, err)
			return
		}
		defer rows.Close()
		res := make([]interface{}, 0, limit)
		for rows.Next() {
			txRow, err := rows.Values()
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			res = append(res, txRow[0])
		}
		c.JSON(200, res)
	})
	router.Run(listenAddr)
}
