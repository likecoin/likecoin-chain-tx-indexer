package parallel

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func MigrateNftOwner(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating NFT owner addresses")
	prevBatchLastPkey := int64(0)
	pkeyIds := make([]int64, batchSize)
	owners := make([]string, batchSize)
	for {
		rows, err := conn.Query(context.Background(), `
			SELECT id, owner
			FROM nft
			WHERE id > $1
			ORDER BY id
			LIMIT $2
		`, prevBatchLastPkey, batchSize)
		if err != nil {
			logger.L.Errorw("Error when querying batch", "error", err)
			return err
		}
		count := 0
		for rows.Next() {
			err = rows.Scan(&pkeyIds[count], &owners[count])
			if err != nil {
				logger.L.Errorw("Error when scanning row", "error", err)
				return err
			}
			count++
		}
		if count == 0 {
			break
		}
		batch := pgx.Batch{}
		for i := 0; i < count; i++ {
			pkeyId := pkeyIds[i]
			convertedOwner, err := utils.ConvertAddressPrefix(owners[i], db.MainAddressPrefix)
			if err != nil {
				logger.L.Warnw(
					"Cannot convert NFT owner address",
					"pkey_id", pkeyId,
					"owner", owners[i],
					"error", err,
				)
				continue
			}
			if convertedOwner != owners[i] {
				batch.Queue(
					`UPDATE nft SET owner = $1 WHERE id = $2`,
					convertedOwner, pkeyId,
				)
			}
		}
		if batch.Len() > 0 {
			results := conn.SendBatch(context.Background(), &batch)
			_, err = results.Exec()
			if err != nil {
				logger.L.Errorw(
					"Error when updating NFT owner",
					"batch_first_pkey_id", pkeyIds[0],
					"batch_last_pkey_id", pkeyIds[count-1],
					"error", err,
				)
				return err
			}
			results.Close()
		}
		prevBatchLastPkey = pkeyIds[count-1]
		logger.L.Infow(
			"NFT owner migration progress",
			"pkey_id", prevBatchLastPkey,
		)
	}
	logger.L.Info("Migration for NFT owner addresses done")
	return nil
}
