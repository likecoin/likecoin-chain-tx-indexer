package db

import (
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func GetNftCount(conn *pgxpool.Conn, q QueryNftCountRequest) (count QueryCountResponse, err error) {
	sql := `
	SELECT COUNT(DISTINCT n.id)
	FROM nft as n
	JOIN nft_class AS c USING (class_id)
	JOIN iscn as i ON c.parent_iscn_id_prefix = i.iscn_id_prefix
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
	// Message 0 is authz MsgExec
	sql := `
	SELECT COUNT(*), sum((tx #>> '{"tx", "body", "messages", 0, "msgs", 0, "amount", 0, "amount"}')::bigint)
	FROM txs
	JOIN (
		SELECT DISTINCT tx_hash FROM nft_event
		WHERE sender = $1
		AND action = '/cosmos.nft.v1beta1.MsgSend'
	) t
	ON tx_hash = tx ->> 'txhash'::text
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	err = conn.QueryRow(ctx, sql, q.ApiAddress).Scan(
		&res.Count, &res.TotalVolume,
	)
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
