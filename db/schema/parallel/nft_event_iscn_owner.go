package parallel

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func MigrateNftEventIscnOwner(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	err = checkMinSchemaVersion(conn, 16)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating NFT event ISCN owner")
	var maxID int64
	row := conn.QueryRow(context.Background(), `SELECT max(id) FROM nft_event`)
	err = row.Scan(&maxID)
	if err != nil {
		logger.L.Errorw("Error when querying max ID", "error", err)
		return err
	}
	logger.L.Debugw("Got max ID", "max_id", maxID)
	for batchHead := int64(0); batchHead <= maxID; batchHead += int64(batchSize) {
		batchUntil := batchHead + int64(batchSize)
		logger.L.Debugf("Batch start, head = %d", batchHead)
		_, err = conn.Exec(context.Background(), `
			UPDATE nft_event AS e
			SET iscn_owner_at_the_time = (
				SELECT DISTINCT i.owner
			)
			FROM
				nft_class AS c,
				iscn AS i,
				iscn_latest_version AS v
			WHERE
				e.class_id = c.class_id
				AND i.iscn_id_prefix = c.parent_iscn_id_prefix
				AND i.iscn_id_prefix = v.iscn_id_prefix
				AND i.version = v.latest_version
				AND e.iscn_owner_at_the_time = ''
				AND e.id >= $1
				AND e.id < $2
		`, batchHead, batchUntil)
		if err != nil {
			logger.L.Errorw(
				"Error when executing UPDATE statement on nft_event table",
				"batch_head", batchHead,
				"max_id", maxID,
				"batch_size", batchSize,
				"error", err,
			)
			return err
		}
		logger.L.Infow(
			"NFT event ISCN owner migration progress",
			"batch_head", batchHead,
			"batch_size", batchSize,
			"max_id", maxID,
		)
	}
	logger.L.Info("Migration for NFT event ISCN owner done")
	return nil
}
