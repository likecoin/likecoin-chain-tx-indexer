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
	accountVariations := utils.ConvertAddressPrefixes(q.Account, AddressPrefixes)
	iscnOwnerVariations := utils.ConvertAddressPrefixes(q.IscnOwner, AddressPrefixes)
	sql := fmt.Sprintf(`
	SELECT DISTINCT ON (c.id)
		c.id, c.class_id, c.name, c.description, c.symbol, c.uri, c.uri_hash,
		c.config, c.metadata, c.price,
		c.parent_type, c.parent_iscn_id_prefix, c.parent_account, c.created_at,
	(
		SELECT array_agg(row_to_json((n.*)))
		FROM nft as n
		WHERE n.class_id = c.class_id
			AND $7 = true
		GROUP BY n.class_id
	) as nfts
	FROM nft_class as c
	LEFT JOIN iscn AS i ON i.iscn_id_prefix = c.parent_iscn_id_prefix
	LEFT JOIN iscn_latest_version
	ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
	WHERE ($4 = '' OR c.parent_iscn_id_prefix = $4)
		AND ($5::text[] IS NULL OR cardinality($5::text[]) = 0 OR c.parent_account = ANY($5))
		AND ($6::text[] IS NULL OR cardinality($6::text[]) = 0 OR i.owner = ANY($6))
		AND ($8 = true OR i.version = iscn_latest_version.latest_version)
		AND ($1 = 0 OR c.id > $1)
		AND ($2 = 0 OR c.id < $2)
	ORDER BY c.id %s
	LIMIT $3
	`, p.Order())
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(
		ctx, sql,
		p.After(), p.Before(), p.Limit, q.IscnIdPrefix, accountVariations,
		iscnOwnerVariations, q.Expand, q.AllIscnVersions)
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

func GetClassesRanking(conn *pgxpool.Conn, q QueryRankingRequest, p PageRequest) (QueryRankingResponse, error) {
	stakeholderIdVariataions := utils.ConvertAddressPrefixes(q.StakeholderId, AddressPrefixes)
	creatorVariations := utils.ConvertAddressPrefixes(q.Creator, AddressPrefixes)
	collectorVariations := utils.ConvertAddressPrefixes(q.Collector, AddressPrefixes)
	ignoreListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreList, AddressPrefixes)
	ApiAddressesVariations := utils.ConvertAddressArrayPrefixes(q.ApiAddresses, AddressPrefixes)
	sql := `
	SELECT
		c.class_id, c.name, c.description, c.symbol, c.uri, c.uri_hash,
		c.config, c.metadata, c.price,
		c.parent_type, c.parent_iscn_id_prefix, c.parent_account, c.created_at,
		COUNT(DISTINCT t.nft_id) AS sold_count,
		SUM(t.price) AS total_sold_value
	FROM (
		SELECT DISTINCT ON (n.id)
			n.nft_id,
			c.id AS class_pid,
			(txs.tx #>> '{"tx", "body", "messages", 0, "msgs", 0, "amount", 0, "amount"}')::bigint AS price
		FROM nft_class AS c
		JOIN nft AS n
			ON c.class_id = n.class_id
		JOIN nft_event AS e
			ON e.nft_id = n.nft_id
		JOIN txs
			ON e.tx_hash = txs.tx ->> 'txhash'
		JOIN iscn AS i
			ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN iscn_latest_version
			ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
				AND i.version = iscn_latest_version.latest_version
		LEFT JOIN iscn_stakeholders
			ON i.id = iscn_pid
		WHERE
			($2 = true OR n.owner != i.owner)
			AND ($3::text[] IS NULL OR cardinality($3::text[]) = 0 OR n.owner != ALL($3))
			AND ($4::text[] IS NULL OR cardinality($4::text[]) = 0 OR i.owner = ANY($4))
			AND ($5 = '' OR i.data #>> '{"contentMetadata", "@type"}' = $5)
			AND ($6::text[] IS NULL OR cardinality($6::text[]) = 0 OR sid = ANY($6))
			AND ($7 = '' OR sname = $7)
			AND ($8::text[] IS NULL OR cardinality($8::text[]) = 0 OR n.owner = ANY($8))
			AND ($9 = 0 OR c.created_at > to_timestamp($9))
			AND ($10 = 0 OR c.created_at < to_timestamp($10))
			AND e.action = '/cosmos.nft.v1beta1.MsgSend'
			AND ($11 = 0 OR (e.timestamp IS NOT NULL AND e.timestamp > to_timestamp($11)))
			AND ($12 = 0 OR (e.timestamp IS NOT NULL AND e.timestamp < to_timestamp($12)))
			AND e.sender = ANY($13::text[])
	) AS t
	JOIN nft_class AS c
		ON c.id = t.class_pid
	GROUP BY c.id
	ORDER BY total_sold_value DESC
	LIMIT $1
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql,
		// $1 ~ $5
		p.Limit, q.IncludeOwner, ignoreListVariations, creatorVariations, q.Type,
		// $6 ~ $10
		stakeholderIdVariataions, q.StakeholderName, collectorVariations, q.CreatedAfter, q.CreatedBefore,
		// $11 ~ 13
		q.After, q.Before, ApiAddressesVariations,
	)
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
			&c.CreatedAt, &c.SoldCount, &c.TotalSoldValue,
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
	ownerVariations := utils.ConvertAddressPrefixes(q.Owner, AddressPrefixes)
	sql := fmt.Sprintf(`
	SELECT n.id, n.nft_id, n.class_id, n.owner, n.uri, n.uri_hash, n.metadata,
		e.timestamp, c.parent_type, c.parent_iscn_id_prefix, c.parent_account
	FROM nft as n
	JOIN nft_class as c
	ON n.class_id = c.class_id
	JOIN (
		SELECT DISTINCT ON (nft_id) nft_id, timestamp
		FROM nft_event
		WHERE receiver = ANY($4)
		ORDER BY nft_id, timestamp DESC
	) e
	ON n.nft_id = e.nft_id
	WHERE owner = ANY($4)
		AND ($1 = 0 OR n.id > $1)
		AND ($2 = 0 OR n.id < $2)
	ORDER BY n.id %s
	LIMIT $3
	`, p.Order())
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, p.After(), p.Before(), p.Limit, ownerVariations)
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
	ignoreFromListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreFromList, AddressPrefixes)
	ignoreToListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreToList, AddressPrefixes)
	sql := fmt.Sprintf(`
	SELECT e.id, action, e.class_id, e.nft_id, e.sender, e.receiver, e.timestamp, e.tx_hash, e.events, t.tx -> 'tx' -> 'body' ->> 'memo' AS memo
	FROM nft_event as e
	JOIN nft_class as c
	ON e.class_id = c.class_id
	JOIN txs AS t
	ON e.tx_hash = t.tx ->> 'txhash'::text
	WHERE ($4 = '' OR e.class_id = $4)
		AND (nft_id = '' OR $5 = '' OR nft_id = $5)
		AND ($6 = '' OR c.parent_iscn_id_prefix = $6)
		AND ($1 = 0 OR e.id > $1)
		AND ($2 = 0 OR e.id < $2)
		AND ($7::text[] IS NULL OR cardinality($7::text[]) = 0 OR e.action = ANY($7))
		AND ($8::text[] IS NULL OR cardinality($8::text[]) = 0 OR e.sender != ALL($8))
		AND ($9::text[] IS NULL OR cardinality($9::text[]) = 0 OR e.receiver != ALL($9))
	ORDER BY e.id %s
	LIMIT $3
	`, p.Order())

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(
		ctx, sql,
		p.After(), p.Before(), p.Limit, q.ClassId, q.NftId,
		q.IscnIdPrefix, q.ActionType, ignoreFromListVariations, ignoreToListVariations,
	)
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
			&e.Receiver, &e.Timestamp, &e.TxHash, &eventRaw, &e.Memo,
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

func GetCollector(conn *pgxpool.Conn, q QueryCollectorRequest, p PageRequest) (res QueryCollectorResponse, err error) {
	creatorVariations := utils.ConvertAddressPrefixes(q.Creator, AddressPrefixes)
	ignoreListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreList, AddressPrefixes)
	sql := `
	SELECT owner, sum(count) as total,
		array_agg(json_build_object(
			'iscn_id_prefix', iscn_id_prefix,
			'class_id', class_id,
			'count', count)),
		COUNT(*) OVER() AS row_count
	FROM (
		SELECT n.owner, i.iscn_id_prefix, c.class_id, COUNT(DISTINCT n.id) as count
		FROM iscn as i
		JOIN iscn_latest_version
		ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
			AND ($5 = true OR i.version = iscn_latest_version.latest_version)
		JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
			AND ($4::text[] IS NULL OR cardinality($4::text[]) = 0 OR n.owner != ALL($4))
		WHERE $1::text[] IS NULL OR cardinality($1::text[]) = 0 OR i.owner = ANY($1)
		GROUP BY n.owner, i.iscn_id_prefix, c.class_id
	) as r
	GROUP BY owner
	ORDER BY total DESC
	OFFSET $2
	LIMIT $3
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, creatorVariations, p.Offset, p.Limit, ignoreListVariations, q.AllIscnVersions)
	if err != nil {
		logger.L.Errorw("failed to query collectors", "error", err, "q", q)
		err = fmt.Errorf("query supporters error: %w", err)
		return
	}
	defer rows.Close()

	res.Collectors, res.Pagination.Total, err = parseAccountCollections(rows)
	if err != nil {
		err = fmt.Errorf("scan collectors error: %w", err)
		return
	}
	res.Pagination.Count = len(res.Collectors)
	return
}

func GetCreators(conn *pgxpool.Conn, q QueryCreatorRequest, p PageRequest) (res QueryCreatorResponse, err error) {
	collectorVariations := utils.ConvertAddressPrefixes(q.Collector, AddressPrefixes)
	ignoreListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreList, AddressPrefixes)
	sql := `
	SELECT owner, sum(count) as total,
		array_agg(json_build_object(
			'iscn_id_prefix', iscn_id_prefix,
			'class_id', class_id,
			'count', count)),
		COUNT(*) OVER() AS row_count
	FROM (
		SELECT i.owner, i.iscn_id_prefix, c.class_id, COUNT(DISTINCT n.id) as count
		FROM iscn as i
		JOIN iscn_latest_version
		ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
			AND ($5 = true OR i.version = iscn_latest_version.latest_version)
		JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
			AND ($4::text[] IS NULL OR cardinality($4::text[]) = 0 OR n.owner != ALL($4))
		WHERE $1::text[] IS NULL OR cardinality($1::text[]) = 0 OR n.owner = ANY($1)
		GROUP BY i.owner, i.iscn_id_prefix, c.class_id
	) as r
	GROUP BY owner
	ORDER BY total DESC
	OFFSET $2
	LIMIT $3
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, collectorVariations, p.Offset, p.Limit, ignoreListVariations, q.AllIscnVersions)
	if err != nil {
		logger.L.Errorw("failed to query creators", "error", err, "q", q)
		err = fmt.Errorf("query creators error: %w", err)
		return
	}

	res.Creators, res.Pagination.Total, err = parseAccountCollections(rows)
	if err != nil {
		err = fmt.Errorf("scan creators error: %w", err)
		return
	}
	res.Pagination.Count = len(res.Creators)
	return
}

func parseAccountCollections(rows pgx.Rows) (accounts []accountCollection, rowCount int, err error) {
	for rows.Next() {
		var account accountCollection
		var collections pgtype.JSONBArray
		if err = rows.Scan(&account.Account, &account.Count, &collections, &rowCount); err != nil {
			return
		}
		if err = collections.AssignTo(&account.Collections); err != nil {
			return
		}
		accounts = append(accounts, account)
	}
	return
}

func GetUserStat(conn *pgxpool.Conn, q QueryUserStatRequest) (res QueryUserStatResponse, err error) {
	res = QueryUserStatResponse{
		CollectedClasses: make([]CollectedClass, 0),
	}

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := `
	SELECT c.class_id, COUNT(c.id)
	FROM nft_class as c
	JOIN nft AS n ON c.class_id = n.class_id
	WHERE n.owner = $1
	GROUP BY c.class_id
	`
	rows, err := conn.Query(ctx, sql, q.User)
	if err != nil {
		logger.L.Errorw("failed to query collected classes", "error", err, "q", q)
		err = fmt.Errorf("query collected classes error: %w", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c CollectedClass
		if err = rows.Scan(&c.ClassId, &c.Count); err != nil {
			err = fmt.Errorf("scan collected classes error: %w", err)
			return
		}
		res.CollectedClasses = append(res.CollectedClasses, c)
	}

	sql = `
	SELECT COUNT(DISTINCT(c.class_id))
	FROM iscn AS i
	JOIN iscn_latest_version
	ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
		AND ($2 = true OR i.version = iscn_latest_version.latest_version)
	JOIN nft_class AS c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
	WHERE i.owner = $1
	`

	row := conn.QueryRow(ctx, sql, q.User, q.AllIscnVersions)

	if err = row.Scan(&res.CreatedCount); err != nil {
		err = fmt.Errorf("scan created count error: %w", err)
		return
	}

	sql = `
	SELECT COUNT(DISTINCT(n.owner))
	FROM iscn AS i
	JOIN iscn_latest_version
	ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
		AND ($3 = true OR i.version = iscn_latest_version.latest_version)
	JOIN nft_class AS c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
	JOIN nft AS n ON c.class_id = n.class_id
		AND ($2::text[] IS NULL OR n.owner != ALL($2))
	WHERE i.owner = $1
	`

	row = conn.QueryRow(ctx, sql, q.User, q.IgnoreList, q.AllIscnVersions)

	err = row.Scan(&res.CollectorCount)
	if err != nil {
		err = fmt.Errorf("scan collector count error: %w", err)
		return
	}

	return
}
