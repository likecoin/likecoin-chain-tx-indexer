package db

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func QueryISCN(conn *pgxpool.Conn, query ISCNQuery, page PageRequest) (ISCNResponse, error) {
	sql := fmt.Sprintf(`
			SELECT DISTINCT ON (id) id, iscn_id, owner, timestamp, ipld, iscn.data
			FROM iscn
			JOIN iscn_stakeholders
			ON id = iscn_pid
			WHERE
				($1 = '' OR iscn_id = $1)
				AND ($2 = '' OR owner = $2)
				AND ($3::text[] IS NULL OR keywords @> $3)
				AND ($4::text[] IS NULL OR fingerprints @> $4)
				AND ($5 = '' OR sid = $5)
				AND ($6 = '' OR sname = $6)
				AND ($7 = 0 OR id > $7)
				AND ($8 = 0 OR id < $8)
			ORDER BY id %s, timestamp
			LIMIT $9;
		`, page.Order())

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(
		ctx, sql,
		query.IscnID, query.Owner, query.Keywords, query.Fingerprints, query.StakeholderID, query.StakeholderName,
		page.After(), page.Before(), page.Limit,
	)
	if err != nil {
		logger.L.Errorw("Query ISCN failed", "error", err, "iscn_query", query)
		return ISCNResponse{}, fmt.Errorf("Query ISCN failed: %w", err)
	}
	defer rows.Close()

	return parseISCN(rows)
}

func QueryISCNList(conn *pgxpool.Conn, pagination PageRequest) (ISCNResponse, error) {
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := fmt.Sprintf(`
		SELECT id, iscn_id, owner, timestamp, ipld, data
		FROM iscn
		WHERE ($1 = 0 OR id > $1)
			AND ($2 = 0 OR id < $2)
		ORDER BY id %s
		LIMIT $3;
	`, pagination.Order())
	rows, err := conn.Query(ctx, sql, pagination.After(), pagination.Before(), pagination.Limit)
	if err != nil {
		logger.L.Errorw("Query error:", "error", err)
		return ISCNResponse{}, err
	}
	defer rows.Close()
	return parseISCN(rows)
}

func QueryISCNAll(conn *pgxpool.Conn, term string, pagination PageRequest) (ISCNResponse, error) {
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
				ORDER BY id %s, timestamp
				LIMIT $5
			)
			UNION ALL
			(
				SELECT DISTINCT ON (id) id, iscn_id, owner, timestamp, ipld, iscn.data
				FROM iscn
				WHERE
					(
						iscn_id = $1
						OR owner = $1
						OR keywords @> $2::text[]
						OR fingerprints @> $2::text[]
					)
					AND ($3 = 0 OR id > $3)
					AND ($4 = 0 OR id < $4)
				ORDER BY id %s, timestamp
				LIMIT $5
			)
		) t
		ORDER BY id %s, timestamp
		LIMIT $5
	`, order, order, order) // mayday, mayday, mayday

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql,
		term, []string{term},
		pagination.After(), pagination.Before(), pagination.Limit,
	)
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
		err := rows.Scan(&res.Pagination.NextKey, &iscn.Id, &iscn.Owner, &iscn.RecordTimestamp, &ipld, &data)
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
	res.Pagination.Count = len(res.Records)
	return res, nil
}
