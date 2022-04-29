package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	iscnTypes "github.com/likecoin/likechain/x/iscn/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

type ISCNResponse struct {
	iscnTypes.IscnRecord
	Id              string `json:"@id"`
	Type            string `json:"@type"`
	RecordTimestamp string `json:"recordTimestamp"`
	Owner           string `json:"owner"`
}

type queryResult struct {
	rows pgx.Rows
	err error
}

func debugSQL(tx pgx.Tx, ctx context.Context, sql string, args ...interface{}) (err error) {
	rows, err := tx.Query(ctx, "EXPLAIN "+sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() && err == nil {
		var line string
		err = rows.Scan(&line)
		log.Println(line)
	}
	return err
}

func QueryISCN(pool *pgxpool.Pool, events types.StringEvents, query ISCNRecordQuery, keywords Keywords, pagination Pagination) ([]iscnTypes.QueryResponseRecord, error) {
	eventStrings := getEventStrings(events)
	queryString, err := query.Marshal()
	if err != nil {
		return nil, err
	}
	keywordString := keywords.Marshal()
	log.Println(eventStrings, string(queryString), keywordString)
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	resultChan := make(chan queryResult, 1)

	go func () {
		conn, err := AcquireFromPool(pool)
		if err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		defer conn.Release()

		txWithIndex, err := conn.Begin(ctx)
		if err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		if _, err := txWithIndex.Exec(ctx, `SET LOCAL enable_indexscan = off;`); err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		
		defer txWithIndex.Rollback(ctx)
		queryISCN(resultChan, ctx, txWithIndex, eventStrings, string(queryString), keywordString, pagination)
		log.Println("Got result WITH index")
	}()
	go func() {
		conn, err := AcquireFromPool(pool)
		if err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		defer conn.Release()

		txWithoutIndex, err := conn.Begin(ctx)
		if err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		queryISCN(resultChan, ctx, txWithoutIndex, eventStrings, string(queryString), keywordString, pagination)
		log.Println("Got result WITHOUT index")
	}()
	
	select {
	case result := <- resultChan:
		if result.err != nil {
			log.Fatal(result.err)
			return nil, result.err
		}
		defer result.rows.Close()
		return parseISCNRecords(result.rows)

	case <- time.After(20 * time.Second):
		return nil, fmt.Errorf("Timeout")
	}
}

func queryISCN(result chan queryResult, ctx context.Context, tx pgx.Tx, eventStrings []string, queryString string, keywordString string, pagination Pagination) {
	// For GIN work with pagination
	sql := fmt.Sprintf(`
		SELECT tx #> '{"tx", "body", "messages", 0, "record"}' as data, events, tx #> '{"timestamp"}'
		FROM txs
		WHERE events @> $1
		AND ($2 = '{}' OR tx #> '{tx, body, messages, 0, record}' @> $2::jsonb)
		AND string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') @> $3
		ORDER BY id %s
		OFFSET $4
		LIMIT $5;
	`, pagination.Order)
	// debugSQL(tx, ctx, sql, eventStrings, queryString, keywordString, pagination.getOffset(), pagination.Limit)
	rows, err := tx.Query(ctx, sql, eventStrings, string(queryString), keywordString,
		pagination.getOffset(), pagination.Limit)
	result <- queryResult{rows, err}
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
			logger.L.Errorw("Failed to unmarshal ISCN body from sql response", "jsonb", jsonb, "error", err)
			return
		}

		var events types.StringEvents
		events, err = parseEvents(eventsRows)
		if err != nil {
			return nil, fmt.Errorf("failed to parse events: %w", err)
		}

		iscn.Id = getEventsValue(events, "iscn_record", "iscn_id")
		iscn.Owner = getEventsValue(events, "iscn_record", "owner")
		iscn.Type = "Record"
		iscn.RecordTimestamp = strings.Trim(timestamp, "\"")

		var data []byte
		if data, err = json.Marshal(iscn); err != nil {
			logger.L.Errorw("Failed to marshal ISCN body to []byte", "iscn", iscn, "error", err)
			return nil, err
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
