package parallel

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/extractor"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func MigrateNftIncome(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	err = checkMinSchemaVersion(conn, 14)
	if err != nil {
		return err
	}
	return conn.BeginFunc(context.Background(), func(dbTx pgx.Tx) error {
		logger.L.Infow("NFT income table migration started")
		// use `WHERE rn = 1` to ensure only the latest event with same tx_hash is processed,
		// that way we can skip most unused `mint_nft` actions mixed in the `/cosmos.nft.v1beta1.MsgSend` actions
		_, err := dbTx.Exec(context.Background(), `
			DECLARE nft_income_migration_cursor CURSOR FOR
				SELECT s.id, s.class_id, s.nft_id, s.tx_hash, txs.tx -> 'logs' AS events
				FROM (
					SELECT e.id, e.class_id, e.nft_id, e.tx_hash, ROW_NUMBER() OVER (PARTITION BY e.tx_hash ORDER BY e.id DESC) AS rn
					FROM nft_event AS e
					WHERE e.action IN ('/cosmos.nft.v1beta1.MsgSend', 'buy_nft', 'sell_nft')
				) AS s
				JOIN txs ON s.tx_hash = txs.tx ->> 'txhash'
				WHERE rn = 1
				ORDER BY id
			;
		`)
		if err != nil {
			logger.L.Errorw("Error when declaring cursor", "error", err)
			return err
		}

		for {
			rows, err := dbTx.Query(context.Background(), fmt.Sprintf(`FETCH %d FROM nft_income_migration_cursor`, batchSize))
			if err != nil {
				logger.L.Errorw("Error when fetching from cursor", "error", err)
				return err
			}

			if rows == nil {
				break
			}

			var pkeyId int64
			var incomes []db.NftIncome
			for rows.Next() {
				var classId string
				var nftId string
				var txHash string
				var eventData pgtype.JSONB
				err = rows.Scan(&pkeyId, &classId, &nftId, &txHash, &eventData)
				if err != nil {
					logger.L.Errorw("Error when scanning row", "error", err)
					return err
				}

				var eventsList db.EventsList
				err = eventData.AssignTo(&eventsList)
				if err != nil {
					logger.L.Errorw("Error when parsing events", "error", err)
					return err
				}

				for i, events := range eventsList {
					msgEvents := events.Events
					msgAction := utils.GetEventsValue(msgEvents, "message", "action")
					rawIncomes := []utils.RawIncome{}
					if msgAction == string(db.ACTION_SEND) {
						rawIncomes = extractor.GetRawIncomeFromSendNftMsg(eventsList, i)
					} else if msgAction == string(db.ACTION_BUY) || msgAction == string(db.ACTION_SELL) {
						rawIncomes = extractor.GetRawIncomeFromNftMarketplaceMsgEvents(msgEvents)
					}
					aggregatedIncome := utils.AggregateRawIncomes(rawIncomes)
					for _, income := range aggregatedIncome {
						incomes = append(incomes, db.NftIncome{
							ClassId:   classId,
							NftId:     nftId,
							TxHash:    txHash,
							Address:   income.Address,
							Amount:    income.Amount,
							IsRoyalty: income.IsRoyalty,
						})
					}
				}
			}
			if pkeyId == 0 {
				break
			}
			count := len(incomes)
			logger.L.Infow(
				"NFT income table migration progress",
				"pkey_id", pkeyId,
				"count", count,
			)
			if count == 0 {
				continue
			}
			for i := 0; i < count; i++ {
				income := incomes[i]
				_, err = dbTx.Exec(context.Background(), `
						INSERT INTO nft_income (class_id, nft_id, tx_hash, address, amount, is_royalty)
						VALUES ($1, $2, $3, $4, $5, $6) 
						ON CONFLICT (class_id, nft_id, tx_hash, address) DO UPDATE
						SET amount = excluded.amount, is_royalty = excluded.is_royalty
					`, income.ClassId, income.NftId, income.TxHash, income.Address, income.Amount, income.IsRoyalty)
				if err != nil {
					logger.L.Errorw("Error when inserting into nft_income", "error", err)
					return err
				}
			}
			lastIncome := incomes[count-1]
			logger.L.Infow(
				"NFT income table migration progress",
				"pkey_id", pkeyId,
				"class_id", lastIncome.ClassId,
				"nft_id", lastIncome.NftId,
				"tx_hash", lastIncome.TxHash,
				"address", lastIncome.Address,
				"amount", lastIncome.Amount,
				"is_royalty", lastIncome.IsRoyalty,
			)
		}
		logger.L.Infow("NFT income table migration completed")
		return nil
	})
}
