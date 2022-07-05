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
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
	iscnTypes "github.com/likecoin/likecoin-chain/v2/x/iscn/types"
)

type ISCNResponse struct {
	iscnTypes.IscnRecord
	Id              string    `json:"id"`
	RecordTimestamp time.Time `json:"recordTimestamp"`
	Owner           string    `json:"owner"`
}

type queryResult struct {
	records []iscnTypes.QueryResponseRecord
	err     error
}

func QueryISCN(pool *pgxpool.Pool, iscn ISCN, pagination Pagination) ([]iscnTypes.QueryResponseRecord, error) {
	sql := fmt.Sprintf(`
			SELECT iscn_id, owner, timestamp, ipld, data
			FROM iscn
			WHERE	($1 = '' OR iscn_id = $1)
				AND ($2 = '' OR owner = $2)
				AND ($3::varchar[] IS NULL OR keywords @> $3)
				AND ($4::varchar[] IS NULL OR fingerprints @> $4)
				AND ($5::jsonb IS NULL OR stakeholders @> $5)
			ORDER BY id %s
			OFFSET $6
			LIMIT $7;
		`, pagination.Order)

	conn, err := AcquireFromPool(pool)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	// debugSQL(conn, ctx, sql, iscn.Iscn, iscn.Owner, iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders,
	// 	pagination.getOffset(), pagination.Limit)

	rows, err := conn.Query(ctx, sql, iscn.Iscn, iscn.Owner, iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders,
		pagination.getOffset(), pagination.Limit)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return parseISCN(rows)
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
		SELECT iscn_id, owner, timestamp, ipld, data
		FROM iscn
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
	return parseISCN(rows)
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

		records, err := parseISCNTxs(rows)
		resultChan <- queryResult{records: records, err: err}
	}
	return doubleQuery(pool, callback)
}

func parseISCNTxs(rows pgx.Rows) (res []iscnTypes.QueryResponseRecord, err error) {
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
		events, err = utils.ParseEvents(eventsRows)
		if err != nil {
			logger.L.Warnw("Failed to parse events of db rows", "id", id, "error", err)
			continue
		}

		iscn.Id = utils.GetEventsValue(events, "iscn_record", "iscn_id")
		iscn.Owner = utils.GetEventsValue(events, "iscn_record", "owner")
		iscn.RecordTimestamp, err = time.Parse(time.RFC3339, strings.Trim(timestamp, "\""))
		if err != nil {
			return
		}

		var data []byte
		if data, err = json.Marshal(iscn); err != nil {
			logger.L.Errorw("Failed to marshal ISCN body to []byte", "iscn", iscn, "error", err)
			return nil, err
		}

		result := iscnTypes.QueryResponseRecord{
			Ipld: utils.GetEventsValue(events, "iscn_record", "ipld"),
			Data: iscnTypes.IscnInput(data),
		}
		res = append(res, result)
	}
	return res, nil
}

func parseISCN(rows pgx.Rows) (res []iscnTypes.QueryResponseRecord, err error) {
	res = make([]iscnTypes.QueryResponseRecord, 0)
	for rows.Next() {
		var iscn ISCNResponse
		var ipld string
		var data pgtype.JSONB
		err = rows.Scan(&iscn.Id, &iscn.Owner, &iscn.RecordTimestamp, &ipld, &data)
		if err != nil {
			log.Fatal(err)
		}

		if err = json.Unmarshal(data.Bytes, &iscn); err != nil {
			log.Fatal(err)
		}

		var output []byte
		if output, err = json.Marshal(iscn); err != nil {
			logger.L.Errorw("Failed to marshal ISCN body to []byte", "iscn", iscn, "error", err)
			return nil, err
		}

		result := iscnTypes.QueryResponseRecord{
			Ipld: ipld,
			Data: iscnTypes.IscnInput(output),
		}
		res = append(res, result)
	}
	return res, nil
}
