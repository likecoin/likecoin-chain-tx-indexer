package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

// todo: move to config
const LIMIT = 10000

type EventHandler func(batch *db.Batch, message []byte, events types.StringEvents, timestamp time.Time) error

var handlers = map[string]EventHandler{
	"create_iscn_record":                           insertISCN,
	"/likechain.iscn.MsgCreateIscnRecord":          insertISCN,
	"update_iscn_record":                           insertISCN,
	"/likechain.iscn.MsgUpdateIscnRecord":          insertISCN,
	"msg_change_iscn_record_ownership":             transferISCN,
	"/likechain.iscn.MsgChangeIscnRecordOwnership": transferISCN,
}

func Run(pool *pgxpool.Pool) chan<- int64 {
	trigger := make(chan int64, 100)
	go func() {
		conn, err := db.AcquireFromPool(pool)
		if err != nil {
			logger.L.Errorw("Failed to acquire connection for ISCN extractor", "error", err)
			return
		}

		logger.L.Info("ISCN extractor started")
		var finished bool
		for {
			if err = conn.Ping(context.Background()); err != nil {
				conn, err = db.AcquireFromPool(pool)
				if err != nil {
					logger.L.Errorw("Failed to acquire connection for ISCN extractor", "error", err)
					time.Sleep(10 * time.Second)
					continue
				}
			}
			finished, err = extractISCN(conn)
			if err != nil {
				logger.L.Errorw("Extract ISCN error", "error", err)
				time.Sleep(10 * time.Second)
				continue
			}
			if finished {
				height := <-trigger
				logger.L.Infof("ISCN extractor: trigger by poller, sync to %d", height)
			}
		}
	}()
	return trigger
}

func extractISCN(conn *pgxpool.Conn) (finished bool, err error) {
	begin, err := db.GetMetaHeight(conn, "iscn")
	if err != nil {
		return false, fmt.Errorf("Failed to get ISCN synchonized height: %w", err)
	}

	end, err := db.GetLatestHeight(conn)
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
	log.Println(begin, end)

	ctx, _ := db.GetTimeoutContext()

	sql := `
	SELECT height, tx #> '{"tx", "body", "messages"}' AS messages, tx -> 'logs' AS logs, tx #> '{"timestamp"}' as timestamp
	FROM txs
	WHERE height >= $1
		AND height < $2
		AND events @> '{"message.module=\"iscn\""}'
	ORDER BY height ASC;
	`

	rows, err := conn.Query(ctx, sql, begin, end)
	defer rows.Close()

	batch := db.NewBatch(conn, LIMIT)

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
			log.Println(action)
			if handler, ok := handlers[action]; ok {
				err = handler(&batch, messages[i], event.Events, timestamp)
				if err != nil {
					panic(err)
				}
			} else {
				logger.L.Warnf("Unknown message action: %s", action)
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
