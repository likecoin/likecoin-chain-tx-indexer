package db

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func GetNftMarketplaceItems(conn *pgxpool.Conn, q QueryNftMarketplaceItemsRequest, p PageRequest) (QueryNftMarketplaceItemsResponse, error) {
	blockTime, err := GetLatestBlockTime(conn)
	if err != nil {
		logger.L.Errorw("Failed to get latest block time", "error", err)
		// non-critical error, just use default (0) as blocktime to return all items, including those expired ones
		blockTime = time.Unix(0, 0)
	}
	after := p.After()
	afterTime := time.Unix(int64(after/1e9), int64(after%1e9)).UTC()
	before := p.Before()
	beforeTime := time.Unix(int64(before/1e9), int64(before%1e9)).UTC()
	sql := fmt.Sprintf(`
		SELECT
			m.type, m.class_id, m.nft_id, m.creator, m.price, m.expiration,
			c.metadata AS class_metadata,
			n.metadata AS nft_metadata
		FROM nft_marketplace m
		LEFT JOIN nft_class c
			ON ($11 = true AND m.class_id = c.class_id)
		LEFT JOIN nft n
			ON ($11 = true AND m.nft_id = n.nft_id)
		WHERE (m.expiration > $1)
			AND ($2::bigint = 0 OR m.expiration > $3)
			AND ($4::bigint = 0 OR m.expiration < $5)
			AND (m.type = $7)
			AND ($8 = '' OR m.class_id = $8)
			AND ($9 = '' OR m.nft_id = $9)
			AND ($10 = '' OR m.creator = $10)
		ORDER BY price %s
		LIMIT $6
	`, p.Order())
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(
		ctx, sql,
		// $1 ~ $7
		blockTime, after, afterTime, before, beforeTime, p.Limit, q.Type,
		// $8 ~ $11
		q.ClassId, q.NftId, q.Creator, q.Expand,
	)
	if err != nil {
		logger.L.Errorw("Failed to query database query for GetMarketplaceItems", "error", err, "q", q)
		return QueryNftMarketplaceItemsResponse{}, fmt.Errorf("failed to query database query for GetMarketplaceItems: %w", err)
	}

	var res QueryNftMarketplaceItemsResponse
	for rows.Next() {
		var item NftMarketplaceItemResponse
		if err = rows.Scan(
			&item.Type, &item.ClassId, &item.NftId, &item.Creator, &item.Price, &item.Expiration,
			&item.ClassMetadata, &item.NftMetadata,
		); err != nil {
			logger.L.Errorw("Failed to scan row into NftMarketplaceItemResponse", "error", err)
			return QueryNftMarketplaceItemsResponse{}, fmt.Errorf("failed to scan row into NftMarketplaceItemResponse: %w", err)
		}
		item.Expiration = item.Expiration.UTC()
		res.Pagination.NextKey = uint64(item.Expiration.UnixNano())
		res.Items = append(res.Items, item)
	}
	res.Pagination.Count = len(res.Items)
	return res, nil
}
