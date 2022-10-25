package rest

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain/v3/app"
)

var encodingConfig = app.MakeEncodingConfig()

func getBool(query url.Values, key string) (bool, error) {
	valueStr := query.Get(key)
	if len(valueStr) == 0 || valueStr == "0" || valueStr == "false" {
		return false, nil
	} else {
		return true, nil
	}
}

func getUint(query url.Values, key string) (uint64, error) {
	valueStr := query.Get(key)
	if valueStr == "" {
		return 0, nil
	} else {
		return strconv.ParseUint(valueStr, 10, 64)
	}
}

func getOffset(query url.Values) (uint64, error) {
	offset, err := getUint(query, "pagination.offset")
	if err != nil {
		return 0, fmt.Errorf("cannot parse pagination.offset to unsigned int: %w", err)
	}
	return offset, nil
}

func getQueryOrder(query url.Values) (db.Order, error) {
	orderByStr := strings.ToUpper(query.Get("order_by"))
	switch orderByStr {
	case "", "ORDER_BY_UNSPECIFIED", "ORDER_BY_ASC", "ASC":
		return db.ORDER_ASC, nil
	case "ORDER_BY_DESC", "DESC":
		return db.ORDER_DESC, nil
	default:
		return "", fmt.Errorf("available values for order_by: ORDER_BY_UNSPECIFIED, ORDER_BY_ASC, ORDER_BY_DESC")
	}
}

func trimSingleQuotes(s string) (string, error) {
	if len(s) < 2 {
		return "", fmt.Errorf("invalid query event value: %s", s)
	}
	if s[0] != '\'' || s[len(s)-1] != '\'' {
		return "", fmt.Errorf("expect query event value missing single quotes: %s", s)
	}
	return s[1 : len(s)-1], nil
}

func getLimit(query url.Values, key string) (uint64, error) {
	limit, err := getUint(query, key)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %v to unsigned int: %w", key, err)
	}
	if limit == 0 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	return limit, nil
}

func getPage(query url.Values, key string) (uint64, error) {
	page, err := getUint(query, key)
	if err != nil {
		return 0, fmt.Errorf("cannot parse page to unsigned int: %w", err)
	}
	if page == 0 {
		page = 1
	}
	return page, nil
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
	return events, nil
}

func getConn(c *gin.Context) *pgxpool.Conn {
	return c.MustGet("conn").(*pgxpool.Conn)
}

func getPool(c *gin.Context) *pgxpool.Pool {
	return c.MustGet("pool").(*pgxpool.Pool)
}

func getPagination(c *gin.Context) (p db.PageRequest, err error) {
	p = db.PageRequest{}
	for _, key := range []string{"pagination.key", "pagination.limit", "pagination.reverse", "pagination.offset"} {
		if c.Query(key) != "" {
			err = c.ShouldBindQuery(&p)
			return p, err
		}
	}
	// there is no `pagination.xxx` query, then we fall back to legacy keys
	// everything is in deafult, so we scan
	legacy := db.LegacyPageRequest{}
	err = c.ShouldBindQuery(&legacy)
	p.Key = legacy.Key
	p.Limit = legacy.Limit
	p.Offset = legacy.Offset
	p.Reverse = legacy.Reverse
	return p, err
}
