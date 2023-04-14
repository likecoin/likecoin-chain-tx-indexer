package utils

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func GetEventStrings(events types.StringEvents) []string {
	eventStrings := []string{}
	for _, event := range events {
		for _, attr := range event.Attributes {
			s := fmt.Sprintf("%s.%s=\"%s\"", event.Type, attr.Key, attr.Value)
			if len(s) < 8100 {
				// Cosmos SDK indeed generate meaninglessly long event strings
				// (e.g. in IBC client update, hex-encoding the whole header)
				// These event strings are useless and can't be handled by Postgres GIN index
				eventStrings = append(eventStrings, s)
			}
		}
	}
	return eventStrings
}

func ParseEvents(query []string) (events types.StringEvents, err error) {
	for _, row := range query {
		arr := strings.SplitN(row, "=", 2)
		k, v := arr[0], strings.Trim(arr[1], "\"")
		if strings.Contains(k, ".") {
			arr := strings.SplitN(k, ".", 2)
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
	if len(events) == 0 {
		return nil, fmt.Errorf("events needed")
	}
	return events, nil
}

func GetEventValue(event *types.StringEvent, key string) string {
	for _, attr := range event.Attributes {
		if attr.Key == key {
			return strings.Trim(attr.Value, "\"")
		}
	}
	return ""
}

func GetEventsValue(events types.StringEvents, t string, key string) string {
	for _, event := range events {
		if event.Type == t {
			value := GetEventValue(&event, key)
			if value != "" {
				return value
			}
		}
	}
	return ""
}

func ParseCoinFromEventString(coinStr string) (uint64, error) {
	coin, err := types.ParseCoinNormalized(coinStr)
	if err != nil {
		return 0, err
	}
	return coin.Amount.Uint64(), nil
}

func GetIncomeMap(events types.StringEvents) map[string]uint64 {
	incomeMap := make(map[string]uint64)
	address := ""
	amount := uint64(0)
	for _, event := range events {
		if event.Type == "coin_received" {
			for _, attr := range event.Attributes {
				if attr.Key == "receiver" {
					address = attr.Value
				}
				if attr.Key == "amount" {
					amountStr := attr.Value
					coin, err := types.ParseCoinNormalized(amountStr)
					if err != nil {
						logger.L.Warnw("Failed to parse income from event", "income_str", amountStr, "error", err)
						address = ""
						continue
					}
					amount = coin.Amount.Uint64()
				}
				if address != "" && amount != 0 {
					incomeMap[address] += amount
					address = ""
					amount = 0
				}
			}
		}
	}
	return incomeMap
}
