package db

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	iscnTypes "github.com/likecoin/likecoin-chain/v2/x/iscn/types"
)

type ISCNResponse struct {
	iscnTypes.IscnRecord
	Id              string    `json:"@id"`
	RecordTimestamp time.Time `json:"recordTimestamp"`
	Owner           string    `json:"owner"`
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
	stakeholderId := fmt.Sprintf(`[{"id": "%s"}]`, term)
	stakeholderName := fmt.Sprintf(`[{"name": "%s"}]`, term)
	sql := fmt.Sprintf(`
		SELECT iscn_id, owner, timestamp, ipld, data
		FROM iscn
		WHERE iscn_id = $1
			OR owner = $1
			OR keywords @> $2
			OR fingerprints @> $2
			OR stakeholders @> $3
			OR stakeholders @> $4
		ORDER BY id %s
		OFFSET $5
		LIMIT $6;
		`, pagination.Order)

	conn, err := AcquireFromPool(pool)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, term, []string{term}, stakeholderId, stakeholderName, pagination.getOffset(), pagination.Limit)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer rows.Close()
	return parseISCN(rows)
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

		if err = json.Unmarshal(data.Bytes, &iscn.IscnRecord); err != nil {
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
