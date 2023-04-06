package parallel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func MigrateNftRoyalty(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	err = checkMinSchemaVersion(conn, 14)
	if err != nil {
		return err
	}
	return conn.BeginFunc(context.Background(), func(dbTx pgx.Tx) error {
		_, err := dbTx.Exec(context.Background(), `
			DECLARE nft_royalty_migration_cursor CURSOR FOR
				SELECT id, class_id, nft_id, tx_hash, events
				FROM nft_event
				ORDER BY id
			;
		`)
		if err != nil {
			logger.L.Errorw("Error when declaring cursor", "error", err)
			return err
		}

		for {
			rows, err := dbTx.Query(context.Background(), fmt.Sprintf(`FETCH %d FROM nft_royalty_migration_cursor`, batchSize))
			if err != nil {
				logger.L.Errorw("Error when fetching from cursor", "error", err)
				return err
			}

			if rows == nil {
				break
			}

			var pkeyId int64
			var royalties []db.NftRoyalty
			for rows.Next() {
				var classId string
				var nftId string
				var txHash string
				var eventRaw []string
				err = rows.Scan(&pkeyId, &classId, &nftId, &txHash, &eventRaw)
				if err != nil {
					logger.L.Errorw("Error when scanning row", "error", err)
					return err
				}

				events, err := utils.ParseEvents(eventRaw)
				if err != nil {
					logger.L.Errorw("Error when parsing events", "error", err)
					return err
				}

				msgAction := utils.GetEventsValue(events, "message", "action")
				if msgAction == "/cosmos.bank.v1beta1.MsgSend" || msgAction == string(db.ACTION_BUY) || msgAction == string(db.ACTION_SELL) {
					royaltyMap := utils.GetRoyaltyMap(events)
					for stakeholder, amount := range royaltyMap {
						royalties = append(royalties, db.NftRoyalty{
							ClassId:     classId,
							NftId:       nftId,
							TxHash:      txHash,
							Stakeholder: stakeholder,
							Royalty:     amount,
						})
					}
				}
			}
			count := len(royalties)
			if count == 0 {
				continue
			}
			for i := 0; i < count; i++ {
				royalty := royalties[i]
				_, err = dbTx.Exec(context.Background(), `
						INSERT INTO nft_royalty (class_id, nft_id, tx_hash, stakeholder_address, royalty)
						VALUES ($1, $2, $3, $4, $5);
					`, royalty.ClassId, royalty.NftId, royalty.TxHash, royalty.Stakeholder, royalty.Royalty)
				if err != nil {
					logger.L.Errorw("Error when inserting into nft_royalty", "error", err)
					return err
				}
			}
			lastRoyalty := royalties[count-1]
			logger.L.Infow(
				"NFT royalty table migration progress",
				"pkey_id", pkeyId,
				"class_id", lastRoyalty.ClassId,
				"nft_id", lastRoyalty.NftId,
				"tx_hash", lastRoyalty.TxHash,
				"stakeholder_address", lastRoyalty.Stakeholder,
				"royalty", lastRoyalty.Royalty,
			)
		}
		return nil
	})
}
