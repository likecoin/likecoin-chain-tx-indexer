package parallel

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func MigrateNftEventSenderAndReceiver(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating NFT event sender and receiver addresses")
	prevBatchLastPkey := int64(0)
	pkeyIds := make([]int64, batchSize)
	senders := make([]string, batchSize)
	receivers := make([]string, batchSize)
	for {
		rows, err := conn.Query(context.Background(), `
			SELECT id, sender, receiver
			FROM nft_event
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
			err = rows.Scan(&pkeyIds[count], &senders[count], &receivers[count])
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
			convertedSender, err := utils.ConvertAddressPrefix(senders[i], db.MainAddressPrefix)
			if err != nil {
				logger.L.Warnw(
					"Cannot convert NFT event sender address",
					"pkey_id", pkeyId,
					"sender", senders[i],
					"error", err,
				)
				convertedSender = senders[i]
			}
			convertedReceiver, err := utils.ConvertAddressPrefix(receivers[i], db.MainAddressPrefix)
			if err != nil {
				logger.L.Warnw(
					"Cannot convert NFT event receiver address",
					"pkey_id", pkeyId,
					"receiver", receivers[i],
					"error", err,
				)
				convertedReceiver = receivers[i]
			}
			if convertedSender != senders[i] || convertedReceiver != receivers[i] {
				batch.Queue(
					`UPDATE nft_event
					 SET
						sender = $1,
						receiver = $2
					WHERE id = $3`,
					convertedSender, convertedReceiver, pkeyId,
				)
			}
		}
		if batch.Len() > 0 {
			results := conn.SendBatch(context.Background(), &batch)
			_, err = results.Exec()
			if err != nil {
				logger.L.Errorw(
					"Error when updating NFT event sender and receiver",
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
			"NFT event sender and receiver migration progress",
			"pkey_id", prevBatchLastPkey,
		)
	}
	logger.L.Info("Migration for NFT event sender and receiver addresses done")
	return nil
}
