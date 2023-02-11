package parallel

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func MigrateNftEventMemo(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating NFT event memo")
	batchHeadId := uint64(0)
	var maxId uint64
	row := conn.QueryRow(context.Background(), `SELECT max(id) FROM nft_event`)
	err = row.Scan(&maxId)
	if err != nil {
		logger.L.Errorw("Error when querying max ID", "error", err)
		return err
	}
	for batchHeadId <= maxId {
		_, err = conn.Exec(context.Background(), `
			UPDATE nft_event AS e
			SET memo = (
				SELECT (txs.tx #>> '{"tx", "body", "memo"}')::text
			)
			FROM txs
			WHERE
				e.id >= $1
				AND e.id < ($1 + $2)
				AND e.tx_hash = txs.tx ->> 'txhash'
				AND e.memo = ''
			;
		`, batchHeadId, batchSize)
		if err != nil {
			logger.L.Errorw(
				"Error when executing UPDATE statement",
				"batch_head_id", batchHeadId,
				"batch_size", batchSize,
				"error", err,
			)
			return err
		}
		batchHeadId += batchSize
		logger.L.Infow(
			"NFT event memo migration progress",
			"migrated_upto_id", batchHeadId,
			"max_id_in_table", maxId,
		)
	}
	logger.L.Info("Migration for NFT event memo done")
	return nil
}
