package db

import (
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func GetClasses(conn *pgxpool.Conn, q QueryClassRequest) (QueryClassResponse, error) {
	sql := `
	SELECT c.class_id, c.name, c.description, c.symbol, c.uri, c.uri_hash,
	c.config, c.metadata, c.price,
	c.parent_type, c.parent_iscn_id_prefix, c.parent_account,
	(
		SELECT array_agg(row_to_json((n.*)))
		FROM nft as n
		WHERE n.class_id = c.class_id
			AND $2 = true
		GROUP BY n.class_id
	) as nfts
	FROM nft_class as c
	WHERE c.parent_iscn_id_prefix = $1
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, q.IscnIdPrefix, q.Expand)
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
			&c.Id, &c.Name, &c.Description, &c.Symbol, &c.URI, &c.URIHash,
			&c.Config, &c.Metadata, &c.Price,
			&c.Parent.Type, &c.Parent.IscnIdPrefix, &c.Parent.Account, &nfts,
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
	return res, nil
}

func GetNftByOwner(conn *pgxpool.Conn, q QueryNftRequest) (QueryNftResponse, error) {
	sql := `
	SELECT n.nft_id, n.class_id, n.uri, n.uri_hash, n.metadata,
		c.parent_type, c.parent_iscn_id_prefix, c.parent_account
	FROM nft as n
	JOIN nft_class as c
	ON n.class_id = c.class_id
	WHERE owner = $1
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, q.Owner)
	if err != nil {
		logger.L.Errorw("Failed to query nft by owner", "error", err, "q", q)
		return QueryNftResponse{}, fmt.Errorf("query nft class error: %w", err)
	}
	res := QueryNftResponse{
		Nfts: make([]NftResponse, 0),
	}
	for rows.Next() {
		var n NftResponse
		if err = rows.Scan(&n.NftId, &n.ClassId, &n.Uri, &n.UriHash, &n.Metadata,
			&n.ClassParent.Type, &n.ClassParent.IscnIdPrefix, &n.ClassParent.Account,
		); err != nil {
			panic(err)
			logger.L.Errorw("failed to scan nft", "error", err, "q", q)
			return QueryNftResponse{}, fmt.Errorf("query nft failed: %w", err)
		}
		res.Nfts = append(res.Nfts, n)
	}
	return res, nil
}

func GetOwnerByClassId(conn *pgxpool.Conn, classId string) (QueryOwnerByClassIdResponse, error) {
	sql := `
	SELECT owner, array_agg(nft_id)
	FROM nft
	WHERE class_id = $1
	GROUP BY owner`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, classId)
	if err != nil {
		logger.L.Errorw("Failed to query owner", "error", err)
		return QueryOwnerByClassIdResponse{}, fmt.Errorf("query owner error: %w", err)
	}

	res := QueryOwnerByClassIdResponse{
		ClassId: classId,
		Owners:  make([]QueryOwnerResponse, 0),
	}
	for rows.Next() {
		var owner QueryOwnerResponse
		if err = rows.Scan(&owner.Owner, &owner.Nfts); err != nil {
			panic(err)
			logger.L.Errorw("failed to scan owner", "error", err)
			return QueryOwnerByClassIdResponse{}, fmt.Errorf("query owner data failed: %w", err)
		}
		owner.Count = len(owner.Nfts)
		res.Owners = append(res.Owners, owner)
	}
	return res, nil
}

func GetNftEvents(conn *pgxpool.Conn, q QueryEventsRequest) (QueryEventsResponse, error) {
	sql := `
	SELECT action, e.class_id, e.nft_id, e.sender, e.receiver, e.timestamp, e.tx_hash, e.events
	FROM nft_event as e
	JOIN nft_class as c
	ON e.class_id = c.class_id
	WHERE ($1 = '' OR e.class_id = $1)
		AND (nft_id = '' OR $2 = '' OR nft_id = $2)
		AND ($3 = '' OR c.parent_iscn_id_prefix = $3)
	ORDER BY e.id`

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, q.ClassId, q.NftId, q.IscnIdPrefix)
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
	res.Count = len(res.Events)
	return res, nil
}
