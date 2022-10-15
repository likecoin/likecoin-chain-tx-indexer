package db

import (
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func GetClasses(conn *pgxpool.Conn, q QueryClassRequest, p PageRequest) (QueryClassResponse, error) {
	sql := fmt.Sprintf(`
	SELECT c.id, c.class_id, c.name, c.description, c.symbol, c.uri, c.uri_hash,
	c.config, c.metadata, c.price,
	c.parent_type, c.parent_iscn_id_prefix, c.parent_account, c.created_at,
	(
		SELECT array_agg(row_to_json((n.*)))
		FROM nft as n
		WHERE n.class_id = c.class_id
			AND $6 = true
		GROUP BY n.class_id
	) as nfts
	FROM nft_class as c
	WHERE ($4 = '' OR c.parent_iscn_id_prefix = $4)
		AND ($5 = '' OR c.parent_account = $5)
		AND ($1 = 0 OR c.id > $1)
		AND ($2 = 0 OR c.id < $2)
	ORDER BY c.id %s
	LIMIT $3
	`, p.Order())
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, p.After(), p.Before(), p.Limit, q.IscnIdPrefix, q.Account, q.Expand)
	if err != nil {
		logger.L.Errorw("Failed to query nft class by iscn id prefix", "error", err, "q", q)
		return QueryClassResponse{}, fmt.Errorf("query nft class by iscn id prefix error: %w", err)
	}

	res := QueryClassResponse{
		Classes: make([]NftClassResponse, 0),
	}
	for rows.Next() {
		var c NftClassResponse
		var nfts pgtype.JSONBArray
		if err = rows.Scan(
			&res.Pagination.NextKey,
			&c.Id, &c.Name, &c.Description, &c.Symbol, &c.URI, &c.URIHash,
			&c.Config, &c.Metadata, &c.Price,
			&c.Parent.Type, &c.Parent.IscnIdPrefix, &c.Parent.Account, &c.CreatedAt, &nfts,
		); err != nil {
			logger.L.Errorw("failed to scan nft class", "error", err)
			return QueryClassResponse{}, fmt.Errorf("query nft class data failed: %w", err)
		}
		if err = nfts.AssignTo(&c.Nfts); err != nil {
			logger.L.Errorw("failed to scan nfts", "error", err)
			return QueryClassResponse{}, fmt.Errorf("query nfts failed: %w", err)
		}
		c.Count = len(c.Nfts)
		res.Classes = append(res.Classes, c)
	}
	res.Pagination.Count = len(res.Classes)
	return res, nil
}

func GetClassesRanking(conn *pgxpool.Conn, q QueryRankingRequest) (QueryRankingResponse, error) {
	sql := `
	SELECT c.class_id, c.name, c.description, c.symbol, c.uri, c.uri_hash,
	c.config, c.metadata, c.price,
	c.parent_type, c.parent_iscn_id_prefix, c.parent_account, c.created_at, sold_count
	FROM nft_class as c
	JOIN (
		SELECT c.id, count(n.id) as sold_count, array_remove(array_agg(DISTINCT n.owner), NULL) as owners
		FROM nft_class as c
		JOIN iscn as i
		ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		LEFT JOIN iscn_stakeholders ON i.id = iscn_pid
		LEFT JOIN nft as n ON c.class_id = n.class_id
			AND ($1 = true OR n.owner != i.owner)
			AND ($2::text[] IS NULL OR n.owner != ALL($2))
		WHERE ($4 = '' OR i.owner = $4)
			AND ($5 = '' OR i.data #>> '{"contentMetadata", "@type"}' = $5)
			AND ($6 = '' OR sid = $6)
			AND ($7 = '' OR sname = $7)
			AND ($9 = 0 OR c.created_at > to_timestamp($9))
			AND ($10 = 0 OR c.created_at < to_timestamp($10))
		GROUP BY c.id
	) AS t USING(id)
	WHERE ($8 = '' OR $8 = ANY(t.owners))
	ORDER BY sold_count DESC
	LIMIT $3
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, q.IncludeOwner, q.IgnoreList,
		q.Limit, q.Creator, q.Type, q.StakeholderId, q.StakeholderName,
		q.Collector, q.After, q.Before)
	if err != nil {
		logger.L.Errorw("Failed to query nft class ranking", "error", err, "q", q)
		return QueryRankingResponse{}, fmt.Errorf("query nft class ranking error: %w", err)
	}

	res := QueryRankingResponse{}
	for rows.Next() {
		var c NftClassRankingResponse
		if err = rows.Scan(
			&c.Id, &c.Name, &c.Description, &c.Symbol, &c.URI, &c.URIHash,
			&c.Config, &c.Metadata, &c.Price,
			&c.Parent.Type, &c.Parent.IscnIdPrefix, &c.Parent.Account,
			&c.CreatedAt, &c.SoldCount,
		); err != nil {
			logger.L.Errorw("failed to scan nft class", "error", err)
			return QueryRankingResponse{}, fmt.Errorf("query nft class data failed: %w", err)
		}
		res.Classes = append(res.Classes, c)
	}
	res.Pagination.Count = len(res.Classes)
	return res, nil
}

func GetNfts(conn *pgxpool.Conn, q QueryNftRequest, p PageRequest) (QueryNftResponse, error) {
	sql := fmt.Sprintf(`
	SELECT n.id, n.nft_id, n.class_id, n.owner, n.uri, n.uri_hash, n.metadata,
		e.timestamp, c.parent_type, c.parent_iscn_id_prefix, c.parent_account
	FROM nft as n
	JOIN nft_class as c
	ON n.class_id = c.class_id
	JOIN (
		SELECT DISTINCT ON (nft_id) nft_id, timestamp
		FROM nft_event
		WHERE receiver = $4
		ORDER BY nft_id, timestamp DESC
	) e
	ON n.nft_id = e.nft_id
	WHERE owner = $4
		AND ($1 = 0 OR n.id > $1)
		AND ($2 = 0 OR n.id < $2)
	ORDER BY n.id %s
	LIMIT $3
	`, p.Order())
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, p.After(), p.Before(), p.Limit, q.Owner)
	if err != nil {
		logger.L.Errorw("Failed to query nft by owner", "error", err, "q", q)
		return QueryNftResponse{}, fmt.Errorf("query nft class error: %w", err)
	}
	res := QueryNftResponse{
		Nfts: make([]NftResponse, 0),
	}
	for rows.Next() {
		var n NftResponse
		if err = rows.Scan(&res.Pagination.NextKey,
			&n.NftId, &n.ClassId, &n.Owner, &n.Uri, &n.UriHash, &n.Metadata,
			&n.Timestamp, &n.ClassParent.Type, &n.ClassParent.IscnIdPrefix, &n.ClassParent.Account,
		); err != nil {
			logger.L.Errorw("failed to scan nft", "error", err, "q", q)
			return QueryNftResponse{}, fmt.Errorf("query nft failed: %w", err)
		}
		res.Nfts = append(res.Nfts, n)
	}
	res.Pagination.Count = len(res.Nfts)
	return res, nil
}

func GetOwners(conn *pgxpool.Conn, q QueryOwnerRequest) (QueryOwnerResponse, error) {
	sql := `
	SELECT owner, array_agg(nft_id)
	FROM nft
	WHERE class_id = $1
	GROUP BY owner
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, q.ClassId)
	if err != nil {
		logger.L.Errorw("Failed to query owner", "error", err)
		return QueryOwnerResponse{}, fmt.Errorf("query owner error: %w", err)
	}

	res := QueryOwnerResponse{
		Owners: make([]OwnerResponse, 0),
	}
	for rows.Next() {
		var owner OwnerResponse
		if err = rows.Scan(&owner.Owner, &owner.Nfts); err != nil {
			logger.L.Errorw("failed to scan owner", "error", err, "q", q)
			return QueryOwnerResponse{}, fmt.Errorf("query owner data failed: %w", err)
		}
		owner.Count = len(owner.Nfts)
		res.Owners = append(res.Owners, owner)
	}
	res.Pagination.Count = len(res.Owners)
	return res, nil
}

func GetNftEvents(conn *pgxpool.Conn, q QueryEventsRequest, p PageRequest) (QueryEventsResponse, error) {
	sql := fmt.Sprintf(`
	SELECT e.id, action, e.class_id, e.nft_id, e.sender, e.receiver, e.timestamp, e.tx_hash, e.events
	FROM nft_event as e
	JOIN nft_class as c
	ON e.class_id = c.class_id
	WHERE ($4 = '' OR e.class_id = $4)
		AND (nft_id = '' OR $5 = '' OR nft_id = $5)
		AND ($6 = '' OR c.parent_iscn_id_prefix = $6)
		AND ($1 = 0 OR e.id > $1)
		AND ($2 = 0 OR e.id < $2)
	ORDER BY e.id %s
	LIMIT $3
	`, p.Order())

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, p.After(), p.Before(), p.Limit, q.ClassId, q.NftId, q.IscnIdPrefix)
	if err != nil {
		logger.L.Errorw("Failed to query nft events", "error", err)
		return QueryEventsResponse{}, fmt.Errorf("query nft events error: %w", err)
	}

	res := QueryEventsResponse{
		Events: make([]NftEvent, 0),
	}
	for rows.Next() {
		var e NftEvent
		var eventRaw []string
		if err = rows.Scan(
			&res.Pagination.NextKey,
			&e.Action, &e.ClassId, &e.NftId, &e.Sender,
			&e.Receiver, &e.Timestamp, &e.TxHash, &eventRaw,
		); err != nil {
			logger.L.Errorw("failed to scan nft events", "error", err, "q", q)
			return QueryEventsResponse{}, fmt.Errorf("query nft events data failed: %w", err)
		}
		if q.Verbose {
			e.Events, err = utils.ParseEvents(eventRaw)
			if err != nil {
				logger.L.Errorw("failed to parse events", "error", err, "event_raw", eventRaw)
				return QueryEventsResponse{}, fmt.Errorf("parse nft events data failed: %w", err)
			}
		}
		res.Events = append(res.Events, e)
	}
	res.Pagination.Count = len(res.Events)
	return res, nil
}

func GetCollector(conn *pgxpool.Conn, q QueryCollectorRequest) (res QueryCollectorResponse, err error) {
	sql := `
	SELECT owner, sum(count) as total,
		array_agg(json_build_object(
			'iscn_id_prefix', iscn_id_prefix,
			'class_id', class_id,
			'count', count))
	FROM (
		SELECT n.owner, i.iscn_id_prefix, c.class_id, COUNT(DISTINCT n.id) as count
		FROM iscn as i
		JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
		WHERE i.owner = $1
		GROUP BY n.owner, i.iscn_id_prefix, c.class_id
	) as r
	GROUP BY owner
	ORDER BY total DESC
	OFFSET $2
	LIMIT $3
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, q.Creator, q.Offset, q.Limit)
	if err != nil {
		logger.L.Errorw("Failed to query collectors", "error", err, "q", q)
		err = fmt.Errorf("query supporters error: %w", err)
		return
	}
	defer rows.Close()

	res.Collectors, err = parseAccountCollections(rows)
	if err != nil {
		err = fmt.Errorf("Scan collectors error: %w", err)
		return
	}
	res.Pagination.Count = len(res.Collectors)
	return
}

func GetCollectorGlobal(conn *pgxpool.Conn, q QueryCollectorRequest) (res QueryCollectorResponse, err error) {
	sql := `
	SELECT owner, sum(count) as total,
		array_agg(json_build_object(
			'iscn_id_prefix', iscn_id_prefix,
			'class_id', class_id,
			'count', count))
	FROM (
		SELECT n.owner, i.iscn_id_prefix, c.class_id, COUNT(DISTINCT n.id) as count
		FROM iscn as i
		JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
		GROUP BY n.owner, i.iscn_id_prefix, c.class_id
	) as r
	GROUP BY owner
	ORDER BY total DESC
	OFFSET $1
	LIMIT $2
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, q.Offset, q.Limit)
	if err != nil {
		logger.L.Errorw("Failed to query collectors", "error", err, "q", q)
		err = fmt.Errorf("query supporters error: %w", err)
		return
	}
	defer rows.Close()

	res.Collectors, err = parseAccountCollections(rows)
	if err != nil {
		err = fmt.Errorf("Scan collectors error: %w", err)
		return
	}
	res.Pagination.Count = len(res.Collectors)
	return
}

func GetCreators(conn *pgxpool.Conn, q QueryCreatorRequest) (res QueryCreatorResponse, err error) {
	sql := `
	SELECT owner, sum(count) as total,
		array_agg(json_build_object(
			'iscn_id_prefix', iscn_id_prefix,
			'class_id', class_id,
			'count', count))
	FROM (
		SELECT i.owner, i.iscn_id_prefix, c.class_id, COUNT(DISTINCT n.id) as count
		FROM iscn as i
		JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
		WHERE n.owner = $1
		GROUP BY i.owner, i.iscn_id_prefix, c.class_id
	) as r
	GROUP BY owner
	ORDER BY total DESC
	OFFSET $2
	LIMIT $3
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, q.Collector, q.Offset, q.Limit)
	if err != nil {
		logger.L.Errorw("Failed to query creators", "error", err, "q", q)
		err = fmt.Errorf("query creators error: %w", err)
		return
	}

	res.Creators, err = parseAccountCollections(rows)
	if err != nil {
		err = fmt.Errorf("Scan creators error: %w", err)
		return
	}
	res.Pagination.Count = len(res.Creators)
	return
}

func GetCreatorsGlobal(conn *pgxpool.Conn, q QueryCreatorRequest) (res QueryCreatorResponse, err error) {
	sql := `
	SELECT owner, sum(count) as total,
		array_agg(json_build_object(
			'iscn_id_prefix', iscn_id_prefix,
			'class_id', class_id,
			'count', count))
	FROM (
		SELECT i.owner, i.iscn_id_prefix, c.class_id, COUNT(DISTINCT n.id) as count
		FROM iscn as i
		JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
		GROUP BY i.owner, i.iscn_id_prefix, c.class_id
	) as r
	GROUP BY owner
	ORDER BY total DESC
	OFFSET $1
	LIMIT $2
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, q.Offset, q.Limit)
	if err != nil {
		logger.L.Errorw("Failed to query creators", "error", err, "q", q)
		err = fmt.Errorf("query creators error: %w", err)
		return
	}

	res.Creators, err = parseAccountCollections(rows)
	if err != nil {
		err = fmt.Errorf("Scan creators error: %w", err)
		return
	}
	res.Pagination.Count = len(res.Creators)
	return
}

func parseAccountCollections(rows pgx.Rows) (accounts []accountCollection, err error) {
	for rows.Next() {
		var account accountCollection
		var collections pgtype.JSONBArray
		if err = rows.Scan(&account.Account, &account.Count, &collections); err != nil {
			return
		}
		if err = collections.AssignTo(&account.Collections); err != nil {
			return
		}
		accounts = append(accounts, account)
	}
	return
}
