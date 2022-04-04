package db

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	iscnTypes "github.com/likecoin/likechain/x/iscn/types"
)

type ISCNResponse struct {
	iscnTypes.IscnRecord
	Id              string `json:"@id"`
	Type            string `json:"@type"`
	RecordTimestamp string `json:"recordTimestamp"`
	Owner           string `json:"owner"`
}

func QueryISCN(conn *pgxpool.Conn, events types.StringEvents, query ISCNRecordQuery, keywords Keywords, pagination Pagination) ([]iscnTypes.QueryResponseRecord, error) {
	eventStrings := getEventStrings(events)
	queryString, err := query.Marshal()
	if err != nil {
		return nil, err
	}
	keywordString := keywords.Marshal()
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// For GIN work with pagination
	if _, err := tx.Exec(ctx, `SET LOCAL enable_indexscan = off;`); err != nil {
		return nil, err
	}
	log.Println(keywordString)
	sql := fmt.Sprintf(`
		SELECT tx #> '{"tx", "body", "messages", 0, "record"}' as data, events, tx #> '{"timestamp"}'
		FROM txs
		WHERE events @> $1 
		AND tx #> '{tx, body, messages, 0, record}' @> $2
		AND string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') @> $3
		ORDER BY id %s
		OFFSET $4
		LIMIT $5;
	`, pagination.Order)
	rows, err := tx.Query(ctx, sql, eventStrings, string(queryString), keywordString,
		pagination.getOffset(), pagination.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseISCNRecords(rows)
}

func QueryISCNList(conn *pgxpool.Conn, pagination Pagination) ([]iscnTypes.QueryResponseRecord, error) {
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := fmt.Sprintf(`
		SELECT tx #> '{"tx", "body", "messages", 0, "record"}' as data, events, tx #> '{"timestamp"}'
		FROM txs
		WHERE events @> '{"message.module=\"iscn\""}'
		ORDER BY id %s
		OFFSET $1
		LIMIT $2;
	`, pagination.Order)
	rows, err := conn.Query(ctx, sql, pagination.getOffset(), pagination.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseISCNRecords(rows)
}

func parseISCNRecords(rows pgx.Rows) (res []iscnTypes.QueryResponseRecord, err error) {
	res = make([]iscnTypes.QueryResponseRecord, 0)
	for rows.Next() && err == nil {
		var iscn ISCNResponse
		var jsonb pgtype.JSONB
		var eventsRows pgtype.VarcharArray
		var timestamp string
		err = rows.Scan(&jsonb, &eventsRows, &timestamp)
		if err != nil {
			return
		}

		if err = json.Unmarshal(jsonb.Bytes, &iscn); err != nil {
			return
		}

		var events types.StringEvents
		events, err = parseEvents(eventsRows)
		if err != nil {
			log.Println(err)
		}

		iscn.Id = getEventsValue(events, "iscn_record", "iscn_id")
		iscn.Owner = getEventsValue(events, "iscn_record", "owner")
		iscn.Type = "Record"
		iscn.RecordTimestamp = strings.Trim(timestamp, "\"")

		var data []byte
		if data, err = json.Marshal(iscn); err != nil {
			return
		}

		result := iscnTypes.QueryResponseRecord{
			Ipld: getEventsValue(events, "iscn_record", "ipld"),
			Data: iscnTypes.IscnInput(data),
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
