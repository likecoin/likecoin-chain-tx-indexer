package db

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

const MAX_LIMIT = 100

func QueryIscn(conn *pgxpool.Conn, query IscnQuery, page PageRequest) (IscnResponse, error) {
	sql := fmt.Sprintf(`
			SELECT DISTINCT ON (id) id, iscn_id, owner, timestamp, ipld, iscn.data
			FROM iscn
			JOIN iscn_stakeholders
			ON id = iscn_pid
			WHERE
				($1 = '' OR iscn_id = $1)
				AND ($2 = '' OR iscn_id_prefix = $2)
				AND ($3 = '' OR owner = $3)
				AND ($4::text[] IS NULL OR keywords @> $4)
				AND ($5::text[] IS NULL OR fingerprints @> $5)
				AND ($6 = '' OR sid = $6)
				AND ($7 = '' OR sname = $7)
				AND ($8 = 0 OR id > $8)
				AND ($9 = 0 OR id < $9)
			ORDER BY id %s, timestamp
			LIMIT %d;
		`, page.Order(), MAX_LIMIT)

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(
		ctx, sql,
		query.IscnId, query.IscnIdPrefix, query.Owner, query.Keywords,
		query.Fingerprints, query.StakeholderId, query.StakeholderName,
		page.After(), page.Before(),
	)
	if err != nil {
		logger.L.Errorw("Query ISCN failed", "error", err, "iscn_query", query)
		return IscnResponse{}, fmt.Errorf("Query ISCN failed: %w", err)
	}
	defer rows.Close()

	return parseIscn(rows, page.Limit)
}

func QueryIscnList(conn *pgxpool.Conn, pagination PageRequest) (IscnResponse, error) {
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := fmt.Sprintf(`
		SELECT id, iscn_id, owner, timestamp, ipld, data
		FROM iscn
		WHERE ($1 = 0 OR id > $1)
			AND ($2 = 0 OR id < $2)
		ORDER BY id %s
		LIMIT %d;
	`, pagination.Order(), MAX_LIMIT)
	rows, err := conn.Query(ctx, sql, pagination.After(), pagination.Before())
	if err != nil {
		logger.L.Errorw("Query error:", "error", err)
		return IscnResponse{}, err
	}
	defer rows.Close()
	return parseIscn(rows, pagination.Limit)
}

func QueryIscnSearch(conn *pgxpool.Conn, term string, pagination PageRequest) (IscnResponse, error) {
	order := pagination.Order()
	sql := fmt.Sprintf(`
		SELECT DISTINCT ON (id) id, iscn_id, owner, timestamp, ipld, data
		FROM (
			(
				SELECT DISTINCT ON (id) id, iscn_id, owner, timestamp, ipld, iscn.data
				FROM iscn
				JOIN iscn_stakeholders
				ON id = iscn_pid
				WHERE
					(
						sid = $1
						OR sname = $1
					)
					AND ($3 = 0 OR id > $3)
					AND ($4 = 0 OR id < $4)
				ORDER BY id, timestamp %[1]s
				LIMIT %[2]d
			)
			UNION ALL
			(
				SELECT DISTINCT ON (id) id, iscn_id, owner, timestamp, ipld, iscn.data
				FROM iscn
				WHERE
					(
						iscn_id = $1
						OR iscn_id_prefix = $1
						OR owner = $1
						OR keywords @> $2::text[]
						OR fingerprints @> $2::text[]
					)
					AND ($3 = 0 OR id > $3)
					AND ($4 = 0 OR id < $4)
				ORDER BY id, timestamp %[1]s
				LIMIT %[2]d
			)
		) t
		ORDER BY id, timestamp %[1]s
		LIMIT %[2]d
	`, order, MAX_LIMIT)

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql,
		term, []string{term},
		pagination.After(), pagination.Before(),
	)
	if err != nil {
		logger.L.Errorw("Query ISCN failed", "error", err, "term", term)
		return IscnResponse{}, fmt.Errorf("Query ISCN failed: %w", err)
	}
	defer rows.Close()
	return parseIscn(rows, pagination.Limit)
}

func parseIscn(rows pgx.Rows, limit int) (IscnResponse, error) {
	res := IscnResponse{}
	for rows.Next() && len(res.Records) < limit {
		var iscn iscnResponseData
		var ipld string
		var data pgtype.JSONB
		err := rows.Scan(&res.Pagination.NextKey, &iscn.Id, &iscn.Owner, &iscn.RecordTimestamp, &ipld, &data)
		if err != nil {
			logger.L.Errorw("Scan ISCN row failed", "error", err)
			return res, fmt.Errorf("Scan ISCN failed: %w", err)
		}

		if err = json.Unmarshal(data.Bytes, &iscn); err != nil {
			logger.L.Errorw("Unmarshal ISCN data failed", "error", err, "data", string(data.Bytes))
			return res, fmt.Errorf("Unmarshal ISCN data failed: %w", err)
		}

		res.Records = append(res.Records, iscnResponseRecord{
			Ipld: ipld,
			Data: iscn,
		})
	}
	res.Pagination.Count = len(res.Records)
	return res, nil
}
