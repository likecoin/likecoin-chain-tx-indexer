package parallel

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func MigrateNftEventPrice(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating NFT event price")
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
			SET price = (
				SELECT (
					txs.tx #>>
						'{"tx", "body", "messages", 0, "msgs", 0, "amount", 0, "amount"}'
				)::bigint
			)
			FROM txs
			WHERE
				e.id >= $1
				AND e.id < ($1 + $2)
				AND e.action = '/cosmos.nft.v1beta1.MsgSend'
				AND e.tx_hash = txs.tx ->> 'txhash'
				AND txs.tx #>> '{"tx", "body", "messages", 0, "@type"}' = '/cosmos.authz.v1beta1.MsgExec'
				AND e.price IS NULL
			;
		`, batchHeadId, batchSize)
		if err != nil {
			logger.L.Errorw(
				"Error when executing UPDATE by MsgExec statement",
				"batch_head_id", batchHeadId,
				"batch_size", batchSize,
				"error", err,
			)
			return err
		}

		_, err = conn.Exec(context.Background(), `
			UPDATE nft_event AS e
			SET price = (
				SELECT (
					txs.tx #>> '{"tx", "body", "messages", 0, "price"}'
				)::bigint
			)
			FROM txs
			WHERE
				e.id >= $1
				AND e.id < ($1 + $2)
				AND e.action = 'buy_nft'
				AND e.tx_hash = txs.tx ->> 'txhash'
				AND txs.tx #>> '{"tx", "body", "messages", 0, "@type"}' = '/likechain.likenft.v1.MsgBuyNFT'
				AND e.price IS NULL
			;
		`, batchHeadId, batchSize)
		if err != nil {
			logger.L.Errorw(
				"Error when executing UPDATE by MsgBuyNFT statement",
				"batch_head_id", batchHeadId,
				"batch_size", batchSize,
				"error", err,
			)
			return err
		}

		batchHeadId += batchSize
		logger.L.Infow(
			"NFT event migration progress",
			"migrated_upto_id", batchHeadId,
			"max_id_in_table", maxId,
		)
	}
	logger.L.Info("Migration for NFT event price done")
	return nil
}
