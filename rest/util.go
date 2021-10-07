package rest

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likechain/app"
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
