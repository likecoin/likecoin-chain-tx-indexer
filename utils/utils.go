package utils

import (
	"strings"
)

func ParseKeywords(keyword string) []string {
	if keyword == "" {
		return []string{}
	}
	arr := strings.Split(keyword, ",")
	res := make([]string, 0, len(arr))
	for _, v := range arr {
		s := strings.Trim(v, " ")
		if s != "" {
			res = append(res, s)
		}
	}
	return res
}

func GuardRankingOrderBy(orderBy string) string {
	switch orderBy {
	case "total_sold_value", "sold_count":
		return orderBy
	default:
		return "total_sold_value"
	}
}
