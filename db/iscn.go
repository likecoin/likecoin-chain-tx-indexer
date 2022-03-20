package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	iscnTypes "github.com/likecoin/likechain/x/iscn/types"
)

func QueryISCN(conn *pgxpool.Conn, events types.StringEvents) ([]iscnTypes.QueryResponseRecord, error) {
	eventStrings := getEventStrings(events)
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	sql := `
		SELECT tx #> '{"tx", "body", "messages", 0, "record"}' as data, events
		FROM txs
		WHERE events @> $1
	`
	rows, err := conn.Query(ctx, sql, eventStrings)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseISCNRecords(rows)
}

func parseISCNRecords(rows pgx.Rows) (res []iscnTypes.QueryResponseRecord, err error) {
	res = make([]iscnTypes.QueryResponseRecord, 0)
	for rows.Next() && err == nil {
		var data pgtype.GenericBinary
		var eventsRows pgtype.VarcharArray
		err = rows.Scan(&data, &eventsRows)
		if err != nil {
			return
		}
		var events types.StringEvents
		events, err = parseEvents(eventsRows)
		if err != nil {
			log.Println(err)
		}

		result := iscnTypes.QueryResponseRecord{
			Ipld: getEventsValue(events, "iscn_record", "ipld"),
			Data: iscnTypes.IscnInput(data.Bytes),
		}
		res = append(res, result)
	}
	return
}

func parseEvents(query pgtype.VarcharArray) (events types.StringEvents, err error) {
	for _, row := range query.Elements {
		arr := strings.SplitN(row.String, "=", 2)
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

func getEventsValue(events types.StringEvents, t string, key string) string {
	for _, event := range events {
		if event.Type == t {
			for _, attr := range event.Attributes {
				if attr.Key == key {
					return attr.Value
				}
			}
		}
	}
	return ""
}
