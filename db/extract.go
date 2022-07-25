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
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

var LIMIT = int64(utils.EnvInt("EXTRACTOR_LIMIT", 10000))

type EventPayload struct {
	Batch     *Batch
	Message   []byte
	Events    types.StringEvents
	Timestamp time.Time
	TxHash    string
}
type EventHandler func(EventPayload) error

func Extract(conn *pgxpool.Conn, handlers map[string]EventHandler) (finished bool, err error) {
	begin, err := GetMetaHeight(conn, META_EXTRACTOR)
	if err != nil {
		return false, fmt.Errorf("Failed to get extractor synchonized height: %w", err)
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
	SELECT tx #> '{"tx", "body", "messages"}' AS messages, tx -> 'logs' AS logs, tx -> 'timestamp', tx -> 'txhash'
	FROM txs
	WHERE height >= $1
		AND height < $2
		AND events && $3
	ORDER BY height ASC;
	`
	eventString := getEventStrings(getHandlingEvents(handlers))

	rows, err := conn.Query(ctx, sql, begin, end, eventString)
	defer rows.Close()

	batch := NewBatch(conn, int(LIMIT))

	for rows.Next() {
		var messageData pgtype.JSONB
		var eventData pgtype.JSONB
		var timestamp time.Time
		var txHash string
		err := rows.Scan(&messageData, &eventData, &timestamp, &txHash)
		if err != nil {
			return false, fmt.Errorf("Failed to scan tx row on tx %s: %w", txHash, err)
		}

		var messages []json.RawMessage
		err = messageData.AssignTo(&messages)
		if err != nil {
			return false, fmt.Errorf("Failed to unmarshal tx message on tx %s: %w", txHash, err)
		}
		var events []struct {
			Events types.StringEvents `json:"events"`
		}

		err = eventData.AssignTo(&events)
		if err != nil {
			return false, fmt.Errorf("Failed to unmarshal tx event on tx %s: %w", txHash, err)
		}

		for i, event := range events {
			action := utils.GetEventsValue(event.Events, "message", "action")
			if handler, ok := handlers[action]; ok {
				payload := EventPayload{
					Batch:     &batch,
					Message:   messages[i],
					Events:    event.Events,
					Timestamp: timestamp,
					TxHash:    strings.Trim(txHash, "\""),
				}
				err = handler(payload)
				if err != nil {
					logger.L.Errorw("Handle message failed", "action", action, "error", err, "payload", payload)
				}
			}
		}
	}
	batch.UpdateMetaHeight(META_EXTRACTOR, end)
	err = batch.Flush()
	if err != nil {
		return false, fmt.Errorf("send batch failed: %w", err)
	}
	logger.L.Infof("Extractor synced height: %d", end)
	return finished, nil
}

func getHandlingEvents(handlers map[string]EventHandler) types.StringEvents {
	result := make(types.StringEvents, 0, len(handlers))
	for action, _ := range handlers {
		result = append(result, types.StringEvent{
			Type: "message",
			Attributes: []types.Attribute{{
				Key:   "action",
				Value: action,
			}},
		})
	}
	return result
}

func GetMetaHeight(conn *pgxpool.Conn, key string) (int64, error) {
	ctx, _ := GetTimeoutContext()
	var height int64
	err := conn.QueryRow(ctx, `SELECT height FROM meta WHERE id = $1`, key).Scan(&height)
	return height, err
}

func (batch *Batch) InsertISCN(iscn ISCN) {
	sql := `
	INSERT INTO iscn (iscn_id, iscn_id_prefix, version, owner, keywords, fingerprints, stakeholders, data, timestamp, ipld, name, description, url) VALUES
	( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	ON CONFLICT DO NOTHING;`
	batch.Batch.Queue(sql, iscn.Iscn, iscn.IscnPrefix, iscn.Version, iscn.Owner,
		iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders, iscn.Data, iscn.Timestamp, iscn.Ipld,
		iscn.Name, iscn.Description, iscn.Url)
}

func (batch *Batch) UpdateMetaHeight(key string, height int64) {
	logger.L.Debugf("Update %s to %d\n", key, height)
	batch.Batch.Queue(`UPDATE meta SET height = $2 WHERE id = $1`, key, height)
}

func (batch *Batch) InsertNftClass(c NftClass) {
	sql := `
	INSERT INTO nft_class
	(class_id, parent_type, parent_iscn_id_prefix, parent_account,
	name, symbol, description, uri, uri_hash,
	metadata, config, price)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	ON CONFLICT DO NOTHING
	`
	batch.Batch.Queue(sql,
		c.Id, c.Parent.Type, c.Parent.IscnIdPrefix, c.Parent.Account,
		c.Name, c.Symbol, c.Description, c.URI, c.URIHash,
		c.Metadata, c.Config, c.Price,
	)
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
}

func (batch *Batch) InsertNft(n Nft) {
	sql := `
	INSERT INTO nft
	(nft_id, class_id, owner, uri, uri_hash, metadata)
	VALUES
	($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING`
	batch.Batch.Queue(sql, n.NftId, n.ClassId, n.Owner, n.Uri, n.UriHash, n.Metadata)
}

func (batch *Batch) InsertNftEvent(e NftEvent) {
	sql := `
	INSERT INTO nft_event
	(action, class_id, nft_id, sender, receiver, events, tx_hash, timestamp)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT DO NOTHING`
	batch.Batch.Queue(sql, e.Action, e.ClassId, e.NftId, e.Sender, e.Receiver, getEventStrings(e.Events), e.TxHash, e.Timestamp)
}
