package parallel

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func MigrateNftPrice(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating NFT latest price")
	var upto int64
	row := conn.QueryRow(context.Background(), `SELECT max(id) FROM nft_event`)
	err = row.Scan(&upto)
	if err != nil {
		logger.L.Errorw("Error when querying max ID", "error", err)
		return err
	}
	for ; upto >= 0; upto -= int64(batchSize) {
		_, err = conn.Exec(context.Background(), `
			UPDATE nft AS n
			SET latest_price = e.price, price_updated_at = e.timestamp
			FROM nft_event AS e
			WHERE
				e.id <= $1
				AND e.id > ($1 - $2)
				AND e.class_id = n.class_id
				AND e.nft_id = n.nft_id
				AND (n.price_updated_at IS NULL OR e.timestamp > n.price_updated_at)
				AND e.price > 0
		`, upto, batchSize)
		if err != nil {
			logger.L.Errorw(
				"Error when executing UPDATE statement on nft table",
				"upto", upto,
				"batch_size", batchSize,
				"error", err,
			)
			return err
		}
		_, err = conn.Exec(context.Background(), `
			UPDATE nft_class AS n
			SET latest_price = e.price, price_updated_at = e.timestamp
			FROM nft_event AS e
			WHERE
				e.id <= $1
				AND e.id > ($1 - $2)
				AND e.class_id = n.class_id
				AND (n.price_updated_at IS NULL OR e.timestamp > n.price_updated_at)
				AND e.price > 0
		`, upto, batchSize)
		if err != nil {
			logger.L.Errorw(
				"Error when executing UPDATE statement on nft_class table",
				"upto", upto,
				"batch_size", batchSize,
				"error", err,
			)
			return err
		}
		logger.L.Infow(
			"NFT latest price migration progress",
			"migrated_upto_id", upto,
		)
	}
	logger.L.Info("Migration for NFT latest price done")
	return nil
}
