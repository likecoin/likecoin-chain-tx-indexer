package db

import (
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
	"github.com/tendermint/tendermint/types/time"
)

func GetISCNOwnerCount(conn *pgxpool.Conn) (count QueryCountResponse, err error) {
	sql := `SELECT COUNT(DISTINCT owner) FROM iscn;`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	err = conn.QueryRow(ctx, sql).Scan(&count.Count)
	if err != nil {
		err = fmt.Errorf("get iscn owner count failed: %w", err)
		logger.L.Error(err)
	}
	return
}

func GetNftCount(conn *pgxpool.Conn, q QueryNftCountRequest) (count QueryCountResponse, err error) {
	sql := `
	SELECT COUNT(DISTINCT n.id)
	FROM nft as n
	JOIN nft_class AS c USING (class_id)
	JOIN iscn as i ON c.parent_iscn_id_prefix = i.iscn_id_prefix
	JOIN iscn_latest_version
	ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
		AND (i.version = iscn_latest_version.latest_version)
	WHERE ($1 = true OR i.owner != n.owner)
		AND ($2::text[] IS NULL OR n.owner != ALL($2))
	`
	ignoreListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreList, AddressPrefixes)
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	err = conn.QueryRow(ctx, sql, q.IncludeOwner, ignoreListVariations).Scan(&count.Count)
	if err != nil {
		err = fmt.Errorf("get nft count failed: %w", err)
		logger.L.Error(err, q)
	}
	return
}

func GetNftTradeStats(conn *pgxpool.Conn, q QueryNftTradeStatsRequest) (res QueryNftTradeStatsResponse, err error) {
	sql := `
	SELECT COUNT(*), SUM(price)
	FROM nft_event
	WHERE price IS NOT NULL AND price > 0
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	err = conn.QueryRow(ctx, sql).Scan(&res.Count, &res.TotalVolume)
	if err != nil {
		err = fmt.Errorf("get nft trade stats failed: %w", err)
		logger.L.Error(err, q)
	}
	return
}

func GetNftCreatorCount(conn *pgxpool.Conn) (count QueryCountResponse, err error) {
	sql := `
	SELECT COUNT(DISTINCT sender) FROM nft_event
	WHERE action = 'new_class';
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	err = conn.QueryRow(ctx, sql).Scan(&count.Count)
	if err != nil {
		err = fmt.Errorf("get nft creator count failed: %w", err)
		logger.L.Error(err)
	}
	return
}

func GetNftReturningCreatorCount(conn *pgxpool.Conn, q QueryNftReturningCreatorCountRequest) (res QueryNftReturningCreatorCountResponse, err error) {
	interval := q.Interval
	if interval == "" {
		interval = "week"
	}

	now := time.Now().Unix()
	oneYear := int64(60 * 60 * 24 * 365.25)
	after, before := q.After, q.Before
	switch {
	case after == 0 && before == 0:
		before = now
		after = now - oneYear
	case after == 0:
		after = before - oneYear
	case before == 0:
		before = after + oneYear
	}

	sql := fmt.Sprintf(`
	WITH user_actions AS (
		SELECT 
			sender,
			timestamp - MIN(timestamp) OVER (PARTITION BY sender) < INTERVAL '%[1]d days' AS is_new,
			DATE_TRUNC('%[2]s', timestamp) AS week_start
		FROM nft_event
		WHERE action = 'new_class'
			AND timestamp BETWEEN to_timestamp($1) AND to_timestamp($2)
	)

	SELECT 
		week_start,
		COUNT(DISTINCT CASE WHEN is_new THEN sender END) AS new_creator_count,
		COUNT(DISTINCT CASE WHEN NOT is_new THEN sender END) AS returning_creator_count
	FROM user_actions
	GROUP BY week_start
	ORDER BY week_start;
	`, q.ReturningThresholdDays, interval)
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, after, before)
	if err != nil {
		err = fmt.Errorf("get nft returning creator count failed: %w", err)
		logger.L.Error(err)
		return res, err
	}

	res = QueryNftReturningCreatorCountResponse{
		Intervals: make([]NftReturningCreatorCountResponse, 0),
	}

	for rows.Next() {
		var week NftReturningCreatorCountResponse
		err = rows.Scan(&week.StartAt, &week.NewCreatorCount, &week.ReturningCreatorCount)
		if err != nil {
			err = fmt.Errorf("scan nft returning creator count failed: %w", err)
			logger.L.Error(err)
			return
		}
		res.Intervals = append(res.Intervals, week)
	}

	return res, nil
}

func GetNftOwnerCount(conn *pgxpool.Conn) (count QueryCountResponse, err error) {
	sql := `
	SELECT COUNT(DISTINCT owner) FROM nft;
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	err = conn.QueryRow(ctx, sql).Scan(&count.Count)
	if err != nil {
		err = fmt.Errorf("get nft owner count failed: %w", err)
		logger.L.Error(err)
	}
	return
}

func GetNftOwnerList(conn *pgxpool.Conn, p PageRequest) (res QueryNftOwnerListResponse, err error) {
	sql := `
	SELECT owner, COUNT(id) FROM nft
	GROUP BY owner
	ORDER BY COUNT(id) DESC
	OFFSET $1
	LIMIT $2
	`

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, p.Offset, p.Limit)
	if err != nil {
		err = fmt.Errorf("get nft owner list failed: %w", err)
		logger.L.Error(err)
		return
	}

	for rows.Next() && err == nil {
		var owner OwnerResponse
		err = rows.Scan(&owner.Owner, &owner.Count)
		res.Owners = append(res.Owners, owner)
	}
	if err != nil {
		err = fmt.Errorf("scan nft owner list failed: %w", err)
		logger.L.Error(err)
	}
	res.Pagination.Count = len(res.Owners)
	return
}
