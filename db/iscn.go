package db

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func QueryISCN(pool *pgxpool.Pool, iscn ISCN, pagination Pagination) (ISCNResponse, error) {
	sql := fmt.Sprintf(`
			SELECT id, iscn_id, owner, timestamp, ipld, data
			FROM iscn
			WHERE	($6 = 0 OR id > $6)
				AND ($7 = 0 OR id < $7)
				AND ($1 = '' OR iscn_id = $1)
				AND ($2 = '' OR owner = $2)
				AND ($3::varchar[] IS NULL OR keywords @> $3)
				AND ($4::varchar[] IS NULL OR fingerprints @> $4)
				AND ($5::jsonb IS NULL OR stakeholders @> $5)
			ORDER BY id %s
			LIMIT $8;
		`, pagination.Order)

	conn, err := AcquireFromPool(pool)
	if err != nil {
		return ISCNResponse{}, err
	}
	defer conn.Release()

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	// debugSQL(conn, ctx, sql, iscn.Iscn, iscn.Owner, iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders,
	// 	pagination.getOffset(), pagination.Limit)

	rows, err := conn.Query(ctx, sql, iscn.Iscn, iscn.Owner, iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders,
		pagination.After, pagination.Before, pagination.Limit)
	if err != nil {
		logger.L.Errorw("Query ISCN failed", "error", err, "iscn query", iscn)
		return ISCNResponse{}, fmt.Errorf("Query ISCN failed: %w", err)
	}
	defer rows.Close()

	return parseISCN(rows)
}

func QueryISCNList(pool *pgxpool.Pool, pagination Pagination) (ISCNResponse, error) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		return ISCNResponse{}, err
	}
	defer conn.Release()

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := fmt.Sprintf(`
		SELECT id, iscn_id, owner, timestamp, ipld, data
		FROM iscn
		WHERE ($1 = 0 OR id > $1)
			AND ($2 = 0 OR id < $2)
		ORDER BY id %s
		LIMIT $3;
	`, pagination.Order)
	rows, err := conn.Query(ctx, sql, pagination.After, pagination.Before, pagination.Limit)
	if err != nil {
		logger.L.Errorw("Query error:", "error", err)
		return ISCNResponse{}, err
	}
	defer rows.Close()
	return parseISCN(rows)
}

func QueryISCNAll(pool *pgxpool.Pool, term string, pagination Pagination) (ISCNResponse, error) {
	stakeholderId := fmt.Sprintf(`[{"id": "%s"}]`, term)
	stakeholderName := fmt.Sprintf(`[{"name": "%s"}]`, term)
	sql := fmt.Sprintf(`
		SELECT id, iscn_id, owner, timestamp, ipld, data
		FROM iscn
		WHERE (iscn_id = $1
			OR owner = $1
			OR keywords @> $2
			OR fingerprints @> $2
			OR stakeholders @> $3
			OR stakeholders @> $4)
			AND ($5 = 0 OR id > $5)
			AND ($6 = 0 OR id < $6)
		ORDER BY id %s
		LIMIT $7
		`, pagination.Order)

	conn, err := AcquireFromPool(pool)
	if err != nil {
		return ISCNResponse{}, err
	}
	defer conn.Release()

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, term, []string{term}, stakeholderId, stakeholderName, pagination.After, pagination.Before, pagination.Limit)
	if err != nil {
		logger.L.Errorw("Query ISCN failed", "error", err, "term", term)
		return ISCNResponse{}, fmt.Errorf("Query ISCN failed: %w", err)
	}
	defer rows.Close()
	return parseISCN(rows)
}

func parseISCN(rows pgx.Rows) (ISCNResponse, error) {
	res := ISCNResponse{}
	for rows.Next() {
		var iscn ISCNResponseData
		var ipld string
		var data pgtype.JSONB
		err := rows.Scan(&res.Last, &iscn.Id, &iscn.Owner, &iscn.RecordTimestamp, &ipld, &data)
		if err != nil {
			logger.L.Errorw("Scan ISCN row failed", "error", err)
			return res, fmt.Errorf("Scan ISCN failed: %w", err)
		}

		if err = json.Unmarshal(data.Bytes, &iscn); err != nil {
			logger.L.Errorw("Unmarshal ISCN data failed", "error", err, "data", string(data.Bytes))
			return res, fmt.Errorf("Unmarshal ISCN data failed: %w", err)
		}

		res.Records = append(res.Records, ISCNResponseRecord{
			Ipld: ipld,
			Data: iscn,
		})
	}
	return res, nil
}
