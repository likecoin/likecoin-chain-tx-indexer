package extractor

import (
	"context"
	"encoding/json"
	"fmt"
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

func RunISCN(pool *pgxpool.Pool, trigger chan int64) {
	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		logger.L.Errorw("Failed to acquire connection for ISCN extractor", "error", err)
		return
	}

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
			select {
			case height := <-trigger:
				logger.L.Infof("ISCN extractor: trigger by poller, sync to %d", height)
			}
		}
	}
	conn.Release()
}

func extractISCN(conn *pgxpool.Conn) (finished bool, err error) {
	begin, err := db.GetISCNHeight(conn)
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

	rows, err := db.GetISCNTxs(conn, begin, end)
	defer rows.Close()

	batch := db.NewBatch(conn, LIMIT)

	for rows.Next() {
		var height int64
		var data pgtype.JSONB
		var eventsRows pgtype.VarcharArray
		var timestamp string
		err = rows.Scan(&height, &data, &eventsRows, &timestamp)
		if err != nil {
			return
		}

		var events types.StringEvents
		events, err = utils.ParseEvents(eventsRows)
		if err != nil {
			logger.L.Warnw("Failed to parse events of db rows", "height", height, "error", err)
			return
		}

		switch action := utils.GetEventsValue(events, "message", "action"); action {
		case "create_iscn_record", "/likechain.iscn.MsgCreateIscnRecord", "update_iscn_record", "/likechain.iscn.MsgUpdateIscnRecord":
			iscn, err := parseISCN(events, data.Bytes, timestamp)
			if err != nil {
				logger.L.Errorw("parse ISCN failed", "error", err, "data", data.Bytes, "events", events)
				break
			}
			batch.InsertISCN(iscn)
		case "msg_change_iscn_record_ownership", "/likechain.iscn.MsgChangeIscnRecordOwnership":
			batch.TransferISCN(events)
		default:
			logger.L.Warnf("Unknown message action: %s", utils.GetEventsValue(events, "message", "action"))
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

func parseISCN(events types.StringEvents, data []byte, timestamp string) (db.ISCN, error) {
	var record db.ISCNRecordQuery
	if err := json.Unmarshal(data, &record); err != nil {
		return db.ISCN{}, fmt.Errorf("Failed to unmarshal iscn: %w", err)
	}
	holders, err := formatStakeholders(record.Stakeholders)
	if err != nil {
		return db.ISCN{}, fmt.Errorf("Failed to format stakeholder, %w", err)
	}
	return db.ISCN{
		Iscn:         utils.GetEventsValue(events, "iscn_record", "iscn_id"),
		IscnPrefix:   utils.GetEventsValue(events, "iscn_record", "iscn_id_prefix"),
		Owner:        utils.GetEventsValue(events, "iscn_record", "owner"),
		Keywords:     utils.ParseKeywords(record.ContentMetadata.Keywords),
		Fingerprints: record.ContentFingerprints,
		Stakeholders: holders,
		Timestamp:    timestamp,
		Ipld:         utils.GetEventsValue(events, "iscn_record", "ipld"),
		Data:         data,
	}, nil
}

func formatStakeholders(stakeholders []db.Stakeholder) ([]byte, error) {
	body := make([]*db.Entity, len(stakeholders))
	for i, v := range stakeholders {
		body[i] = v.Entity
	}
	return json.Marshal(body)
}
