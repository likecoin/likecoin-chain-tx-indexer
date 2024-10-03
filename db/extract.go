package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/pubsub"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

var LIMIT = int64(utils.EnvInt("EXTRACTOR_LIMIT", 10000))

type EventsList []struct {
	Events types.StringEvents `json:"events"`
}

type EventContext struct {
	Batch      *Batch
	Messages   []json.RawMessage
	EventsList EventsList
	Timestamp  time.Time
	TxHash     string
	Memo       string

	// If the event is from authz, we process it by making a psuedo EventContext
	// for each authz message, and then set this field to the original EventContext
	AuthzParent *EventContext
}

type Extractor func(ctx EventContext) error

func Extract(conn *pgxpool.Conn, extractor Extractor) (finished bool, err error) {
	prevSyncedHeight, err := GetMetaHeight(conn, META_EXTRACTOR)
	if err != nil {
		return false, fmt.Errorf("failed to get extractor synchonized height: %w", err)
	}

	latestSyncingHeight, err := GetLatestHeight(conn)
	if err != nil {
		return false, fmt.Errorf("failed to get latest height: %w", err)
	}

	if prevSyncedHeight == latestSyncingHeight {
		return true, nil
	}
	if latestSyncingHeight > prevSyncedHeight+LIMIT {
		latestSyncingHeight = prevSyncedHeight + LIMIT
	} else {
		finished = true
	}

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := `
	SELECT tx #> '{"tx", "body", "messages"}' AS messages, tx -> 'logs' AS logs, tx -> 'timestamp', tx -> 'txhash', tx -> 'tx' -> 'body' -> 'memo'
	FROM txs
	WHERE height > $1
		AND height <= $2
	ORDER BY height ASC;
	`

	rows, err := conn.Query(ctx, sql, prevSyncedHeight, latestSyncingHeight)
	if err != nil {
		return false, fmt.Errorf("failed to query unprocessed txs: %w", err)

	}
	defer rows.Close()

	batch := NewBatch(conn, int(LIMIT))

	for rows.Next() {
		var messageData pgtype.JSONB
		var eventData pgtype.JSONB
		var timestamp time.Time
		var txHash string
		var memo string
		err := rows.Scan(&messageData, &eventData, &timestamp, &txHash, &memo)
		if err != nil {
			return false, fmt.Errorf("failed to scan tx row on tx %s: %w", txHash, err)
		}

		var messages []json.RawMessage
		err = messageData.AssignTo(&messages)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal tx message on tx %s: %w", txHash, err)
		}
		var eventsList EventsList
		err = eventData.AssignTo(&eventsList)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal tx event on tx %s: %w", txHash, err)
		}

		ctx := EventContext{
			Batch:      &batch,
			Messages:   messages,
			EventsList: eventsList,
			Timestamp:  timestamp,
			TxHash:     strings.Trim(txHash, "\""),
			Memo:       strings.Trim(memo, "\""),
		}
		err = extractor(ctx)
		if err != nil {
			logger.L.Errorw("Handle message failed", "error", err, "context", ctx)
		}
	}
	batch.UpdateMetaHeight(META_EXTRACTOR, latestSyncingHeight)
	err = batch.Flush()
	if err != nil {
		return false, fmt.Errorf("send batch failed: %w", err)
	}
	logger.L.Infof("Extractor synced height: %d", latestSyncingHeight)
	return finished, nil
}

func GetMetaHeight(conn *pgxpool.Conn, key string) (int64, error) {
	ctx, _ := GetTimeoutContext()
	var height int64
	err := conn.QueryRow(ctx, `SELECT height FROM meta WHERE id = $1`, key).Scan(&height)
	return height, err
}

func (batch *Batch) InsertIscn(insert IscnInsert) {
	stakeholderIDs := []string{}
	stakeholderNames := []string{}
	stakeholderRawJSONs := [][]byte{}
	for _, s := range insert.Stakeholders {
		stakeholderIDs = append(stakeholderIDs, s.Entity.Id)
		stakeholderNames = append(stakeholderNames, s.Entity.Name)
		stakeholderRawJSONs = append(stakeholderRawJSONs, s.Data)
	}
	convertedOwner, err := utils.ConvertAddressPrefix(insert.Owner, MainAddressPrefix)
	if err == nil {
		insert.Owner = convertedOwner
	}
	sql := `
	WITH result AS (
		INSERT INTO iscn
		(
			iscn_id, iscn_id_prefix, version, owner, keywords,
			fingerprints, data, timestamp, ipld, name,
			description, url
		)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT DO NOTHING
		RETURNING id
	)
	INSERT INTO iscn_stakeholders (iscn_pid, sid, sname, data)
	SELECT id, unnest($13::text[]), unnest($14::text[]), unnest($15::jsonb[])
	FROM result;
	`
	batch.Batch.Queue(sql,
		// $1 ~ $5
		insert.Iscn, insert.IscnPrefix, insert.Version, insert.Owner, insert.Keywords,
		// $6 ~ $10
		insert.Fingerprints, insert.Data, insert.Timestamp, insert.Ipld, insert.Name,
		// $11 ~ $15
		insert.Description, insert.Url, stakeholderIDs, stakeholderNames, stakeholderRawJSONs,
	)
	sql = `
		INSERT INTO iscn_latest_version AS t (iscn_id_prefix, latest_version)
		VALUES ($1, $2)
		ON CONFLICT (iscn_id_prefix) DO UPDATE
			SET latest_version = GREATEST(t.latest_version, EXCLUDED.latest_version)
		;
	`
	batch.Batch.Queue(sql, insert.IscnPrefix, insert.Version)
	_ = pubsub.Publish("NewISCN", insert)
}

func (batch *Batch) UpdateMetaHeight(key string, height int64) {
	logger.L.Debugf("Update %s to %d\n", key, height)
	batch.Batch.Queue(`UPDATE meta SET height = $2 WHERE id = $1`, key, height)
}

func (batch *Batch) InsertNftClass(c NftClass) {
	sql := `
	INSERT INTO nft_class (
		class_id, parent_type, parent_iscn_id_prefix, parent_account, name,
		symbol, description, uri, uri_hash, metadata,
		config, created_at, latest_price, price_updated_at
	)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT DO NOTHING
	`
	batch.Batch.Queue(sql,
		c.Id, c.Parent.Type, c.Parent.IscnIdPrefix, c.Parent.Account, c.Name,
		c.Symbol, c.Description, c.URI, c.URIHash, c.Metadata,
		c.Config, c.CreatedAt, c.LatestPrice, c.PriceUpdatedAt,
	)
	_ = pubsub.Publish("NewNFTClass", c)
}

func (batch *Batch) UpdateNftClass(c NftClass) {
	sql := `
	UPDATE nft_class
	SET name = $1, 
		symbol = $2,
		description = $3, 
		uri = $4,
		uri_hash = $5,
		metadata = $6,
		config = $7
	WHERE class_id = $8
	`
	batch.Batch.Queue(sql,
		c.Name, c.Symbol, c.Description, c.URI, c.URIHash,
		c.Metadata, c.Config, c.Id,
	)
	_ = pubsub.Publish("UpdateNFTClass", c)
}

func (batch *Batch) InsertNft(n Nft) {
	convertedOwner, err := utils.ConvertAddressPrefix(n.Owner, MainAddressPrefix)
	if err == nil {
		n.Owner = convertedOwner
	}
	sql := `
	INSERT INTO nft
	(nft_id, class_id, owner, uri, uri_hash, metadata)
	VALUES
	($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING`
	batch.Batch.Queue(sql, n.NftId, n.ClassId, n.Owner, n.Uri, n.UriHash, n.Metadata)
	_ = pubsub.Publish("NewNFT", n)
}

func (batch *Batch) InsertNftEvent(e NftEvent) {
	convertedSender, err := utils.ConvertAddressPrefix(e.Sender, MainAddressPrefix)
	if err == nil {
		e.Sender = convertedSender
	}
	convertedReceiver, err := utils.ConvertAddressPrefix(e.Receiver, MainAddressPrefix)
	if err == nil {
		e.Receiver = convertedReceiver
	}
	sql := `
	INSERT INTO nft_event (
		action, class_id, nft_id, sender, receiver,
		events, tx_hash, timestamp, price, memo, msg_index,
		iscn_owner_at_the_time
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
		COALESCE(
			(SELECT i.owner
			FROM nft_class AS c
			JOIN iscn AS i
				ON i.iscn_id_prefix = c.parent_iscn_id_prefix
			JOIN iscn_latest_version AS v
				ON i.iscn_id_prefix = v.iscn_id_prefix
					AND i.version = v.latest_version
			WHERE c.class_id = $2
			LIMIT 1)
		, '')
	)
	ON CONFLICT DO NOTHING`
	batch.Batch.Queue(sql,
		e.Action, e.ClassId, e.NftId, e.Sender, e.Receiver,
		utils.GetEventStrings(e.Events), e.TxHash, e.Timestamp, e.Price, e.Memo, e.MsgIndex,
	)

	if e.Price > 0 {
		nftSql := `
			UPDATE nft
			SET latest_price = $1,
				price_updated_at = $2
			WHERE
				class_id = $3
				AND nft_id = $4
		`
		batch.Batch.Queue(nftSql, e.Price, e.Timestamp, e.ClassId, e.NftId)
		nftClassSql := `
			UPDATE nft_class
			SET latest_price = $1,
				price_updated_at = $2
			WHERE
				class_id = $3
		`
		batch.Batch.Queue(nftClassSql, e.Price, e.Timestamp, e.ClassId)
	}
	_ = pubsub.Publish("NewNFTEvent", e)
}

func (batch *Batch) InsertNFTMarketplaceItem(item NftMarketplaceItem) {
	sql := `
	INSERT INTO nft_marketplace (type, class_id, nft_id, creator, price, expiration)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (type, class_id, nft_id, creator) DO UPDATE SET
		price = EXCLUDED.price,
		expiration = EXCLUDED.expiration
	`
	batch.Batch.Queue(sql, item.Type, item.ClassId, item.NftId, item.Creator, item.Price, item.Expiration)
	_ = pubsub.Publish("NewNFTMarketplaceItem", item)
}

func (batch *Batch) InsertNftIncome(income NftIncome) {
	sql := `
	INSERT INTO nft_income (class_id, nft_id, tx_hash, address, amount, is_royalty)
	VALUES ($1, $2, $3, $4, $5, $6)
	`
	batch.Batch.Queue(sql, income.ClassId, income.NftId, income.TxHash, income.Address, income.Amount, income.IsRoyalty)
	_ = pubsub.Publish("NewNFTIncome", income)
}

func (batch *Batch) DeleteNFTMarketplaceItemSilently(item NftMarketplaceItem) {
	sql := `
	DELETE FROM nft_marketplace
	WHERE
		type = $1 AND
		class_id = $2 AND
		nft_id = $3
	`
	batch.Batch.Queue(sql, item.Type, item.ClassId, item.NftId)
}

func (batch *Batch) DeleteNFTMarketplaceItem(item NftMarketplaceItem) {
	batch.DeleteNFTMarketplaceItemSilently(item)
	_ = pubsub.Publish("DeleteNFTMarketplaceItem", item)
}
