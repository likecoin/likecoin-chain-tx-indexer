package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	iscnTypes "github.com/likecoin/likecoin-chain/v3/x/iscn/types"
)

type ISCNResponse struct {
	iscnTypes.IscnRecord
	Id              string `json:"@id"`
	Type            string `json:"@type"`
	RecordTimestamp string `json:"recordTimestamp"`
	Owner           string `json:"owner"`
}

type queryResult struct {
	records []iscnTypes.QueryResponseRecord
	err     error
}

func QueryISCN(pool *pgxpool.Pool, events types.StringEvents, query ISCNRecordQuery, keywords Keywords, pagination Pagination) ([]iscnTypes.QueryResponseRecord, error) {
	eventStrings := getEventStrings(events)
	queryString, err := query.Marshal()
	if err != nil {
		return nil, err
	}
	keywordString := keywords.Marshal()
	callback := func(resultChan chan queryResult, ctx context.Context, tx pgx.Tx) {
		sql := fmt.Sprintf(`
			SELECT id, tx #> '{"tx", "body", "messages", 0, "record"}' as data, events, tx #> '{"timestamp"}'
			FROM txs
			WHERE events @> $1
			AND ($2 = '{}' OR tx #> '{tx, body, messages, 0, record}' @> $2::jsonb)
			AND string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') @> $3
			ORDER BY id %s
			OFFSET $4
			LIMIT $5;
		`, pagination.Order)

		rows, err := tx.Query(ctx, sql, eventStrings, string(queryString), keywordString,
			pagination.getOffset(), pagination.Limit)
		if err != nil {
			resultChan <- queryResult{err: err}
		}
		defer rows.Close()

		records, err := parseISCNRecords(rows)
		resultChan <- queryResult{records: records, err: err}
	}
	return doubleQuery(pool, callback)
}

func doubleQuery(pool *pgxpool.Pool, callback func(chan queryResult, context.Context, pgx.Tx)) ([]iscnTypes.QueryResponseRecord, error) {
	ctx1, cancel1 := GetTimeoutContext()
	ctx2, cancel2 := GetTimeoutContext()
	defer cancel1()
	defer cancel2()

	resultChan := make(chan queryResult, 1)

	go func() {
		conn, err := AcquireFromPool(pool)
		if err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		defer conn.Release()

		txWithIndex, err := conn.Begin(ctx1)
		if err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		// For GIN work with pagination
		if _, err := txWithIndex.Exec(ctx1, `SET LOCAL enable_indexscan = off;`); err != nil {
			resultChan <- queryResult{nil, err}
			return
		}

		defer txWithIndex.Rollback(ctx1)
		callback(resultChan, ctx1, txWithIndex)
	}()
	go func() {
		conn, err := AcquireFromPool(pool)
		if err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		defer conn.Release()

		txWithoutIndex, err := conn.Begin(ctx2)
		if err != nil {
			resultChan <- queryResult{nil, err}
			return
		}
		callback(resultChan, ctx2, txWithoutIndex)
	}()

	select {
	case result := <-resultChan:
		return result.records, result.err

	case <-time.After(45 * time.Second):
		return nil, fmt.Errorf("database query timeout")
	}

}

func QueryISCNList(pool *pgxpool.Pool, pagination Pagination) ([]iscnTypes.QueryResponseRecord, error) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := fmt.Sprintf(`
		SELECT id, tx #> '{"tx", "body", "messages", 0, "record"}' as data, events, tx #> '{"timestamp"}'
		FROM txs
		WHERE events && '{"message.action=\"create_iscn_record\"",
			"message.action=\"update_iscn_record\"",
			"message.action=\"/likechain.iscn.MsgCreateIscnRecord\"",
			"message.action=\"/likechain.iscn.MsgUpdateIscnRecord\""
		}'
		ORDER BY id %s
		OFFSET $1
		LIMIT $2;
	`, pagination.Order)
	rows, err := conn.Query(ctx, sql, pagination.getOffset(), pagination.Limit)
	if err != nil {
		logger.L.Errorw("Query error:", "error", err)
		return nil, err
	}
	defer rows.Close()
	return parseISCNRecords(rows)
}

func QueryISCNAll(pool *pgxpool.Pool, term string, pagination Pagination) ([]iscnTypes.QueryResponseRecord, error) {
	iscnId := fmt.Sprintf(`{"iscn_record.iscn_id=\"%[1]s\""}`, term)
	owner := fmt.Sprintf(`{"iscn_record.owner=\"%[1]s\""}`, term)
	stakeholderId := fmt.Sprintf(`{"stakeholders": [{"entity": {"@id": "%s"}}]}`, term)
	stakeholderName := fmt.Sprintf(`{"stakeholders": [{"entity": {"name": "%s"}}]}`, term)
	contentFingerprints := fmt.Sprintf(`{"contentFingerprints": ["%s"]}`, term)
	keywords := fmt.Sprintf(`{"%s"}`, term)
	sql := fmt.Sprintf(`
			SELECT id, tx #> '{"tx", "body", "messages", 0, "record"}' as data, events, tx #> '{"timestamp"}'
			FROM txs
			WHERE events @> $1
					OR events @> $2
					OR tx #> '{tx, body, messages, 0, record}' @> $3
					OR tx #> '{tx, body, messages, 0, record}' @> $4
					OR tx #> '{tx, body, messages, 0, record}' @> $5
					OR string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') @> $6
			ORDER BY id %s
			OFFSET $7
			LIMIT $8;
		`, pagination.Order)

	callback := func(resultChan chan queryResult, ctx context.Context, tx pgx.Tx) {
		rows, err := tx.Query(ctx, sql, iscnId, owner, stakeholderId, stakeholderName, contentFingerprints, keywords, pagination.getOffset(), pagination.Limit)
		if err != nil {
			resultChan <- queryResult{err: err}
			return
		}
		defer rows.Close()

		records, err := parseISCNRecords(rows)
		resultChan <- queryResult{records: records, err: err}
	}
	return doubleQuery(pool, callback)
}

func parseISCNRecords(rows pgx.Rows) (res []iscnTypes.QueryResponseRecord, err error) {
	res = make([]iscnTypes.QueryResponseRecord, 0)
	for rows.Next() {
		var id uint64
		var iscn ISCNResponse
		var jsonb pgtype.JSONB
		var eventsRows pgtype.VarcharArray
		var timestamp string
		err = rows.Scan(&id, &jsonb, &eventsRows, &timestamp)
		if err != nil {
			return
		}

		if err = json.Unmarshal(jsonb.Bytes, &iscn); err != nil {
			logger.L.Warnw("Failed to unmarshal ISCN body from sql response", "id", id, "error", err)
			continue
		}

		var events types.StringEvents
		events, err = parseEvents(eventsRows)
		if err != nil {
			logger.L.Warnw("Failed to parse events of db rows", "id", id, "error", err)
			continue
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

		iscn.Id = getEventsValue(events, "iscn_record", "iscn_id")
		iscn.Owner = getEventsValue(events, "iscn_record", "owner")
		iscn.Type = "Record"
		iscn.RecordTimestamp = strings.Trim(timestamp, "\"")

		result := iscnTypes.QueryResponseRecord{
			Ipld: getEventsValue(events, "iscn_record", "ipld"),
			Data: iscnTypes.IscnInput(data),
		}
		res = append(res, result)
	}
	return res, nil
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
