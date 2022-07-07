package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

// todo: move to config
const LIMIT = 10000

type EventHandler func(batch *Batch, message []byte, events types.StringEvents, timestamp time.Time) error

func Extract(conn *pgxpool.Conn, handlers map[string]EventHandler) (finished bool, err error) {
	begin, err := GetMetaHeight(conn, "iscn")
	if err != nil {
		return false, fmt.Errorf("Failed to get ISCN synchonized height: %w", err)
	}

	end, err := GetLatestHeight(conn)
	if err != nil {
		return false, fmt.Errorf("Failed to get latest height: %w", err)
	}
	if begin == end {
		return true, nil
	}
	if begin+LIMIT < end {
		end = begin + LIMIT
	} else {
		finished = true
	}

	ctx, _ := GetTimeoutContext()

	sql := `
	SELECT height, tx #> '{"tx", "body", "messages"}' AS messages, tx -> 'logs' AS logs, tx #> '{"timestamp"}' as timestamp
	FROM txs
	WHERE height >= $1
		AND height < $2
	ORDER BY height ASC;
	`

	rows, err := conn.Query(ctx, sql, begin, end)
	defer rows.Close()

	batch := NewBatch(conn, LIMIT)

	for rows.Next() {
		var height int64
		var messageData pgtype.JSONB
		var eventData pgtype.JSONB
		var timestamp time.Time
		err := rows.Scan(&height, &messageData, &eventData, &timestamp)
		if err != nil {
			panic(err)
		}

		var messages []json.RawMessage
		err = messageData.AssignTo(&messages)
		if err != nil {
			panic(err)
		}
		var events []struct {
			Events types.StringEvents `json:"events"`
		}

		err = eventData.AssignTo(&events)
		if err != nil {
			panic(err)
		}

		for i, event := range events {
			action := utils.GetEventsValue(event.Events, "message", "action")
			if handler, ok := handlers[action]; ok {
				err = handler(&batch, messages[i], event.Events, timestamp)
				if err != nil {
					panic(err)
				}
			}
		}
	}
	batch.UpdateISCNHeight(end)
	logger.L.Infof("ISCN synced height: %d", end)
	err = batch.Flush()
	if err != nil {
		return false, fmt.Errorf("send batch failed: %w", err)
	}
	return finished, nil
}

func GetMetaHeight(conn *pgxpool.Conn, key string) (int64, error) {
	ctx, _ := GetTimeoutContext()
	var height int64
	err := conn.QueryRow(ctx, `SELECT height FROM meta WHERE id = $1`, key).Scan(&height)
	return height, err
}

func (batch *Batch) InsertISCN(iscn ISCN) {
	sql := `
INSERT INTO iscn (iscn_id, iscn_id_prefix, version, owner, keywords, fingerprints, stakeholders, data, timestamp, ipld) VALUES
( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT DO NOTHING;`
	batch.Batch.Queue(sql, iscn.Iscn, iscn.IscnPrefix, iscn.Version, iscn.Owner,
		iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders, iscn.Data, iscn.Timestamp, iscn.Ipld)
}

func (batch *Batch) UpdateISCNHeight(height int64) {
	batch.Batch.Queue(`UPDATE meta SET height = $1 WHERE id = 'iscn'`, height)
}

func (batch *Batch) InsertNftClass(c NftClass) {
	sql := `
	INSERT INTO nft_class
	(id, parent_type, parent_iscn_id_prefix, parent_account,
	name, symbol, description, uri, uri_hash,
	metadata, config, price)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	batch.Batch.Queue(sql,
		c.Id, c.Parent.Type, c.Parent.IscnIdPrefix, c.Parent.Account,
		c.Name, c.Symbol, c.Description, c.URI, c.URIHash,
		c.Metadata, c.Config, c.Price,
	)
}

func (batch *Batch) InsertNft(n Nft) {
	sql := `
	INSERT INTO nft
	(nft_id, class_id, owner, uri, uri_hash, metadata)
	VALUES
	($1, $2, $3, $4, $5, $6)`
	batch.Batch.Queue(sql, n.NftId, n.ClassId, n.Owner, n.Uri, n.UriHash, n.Metadata)
}
