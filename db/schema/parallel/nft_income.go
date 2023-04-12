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
		_, err := dbTx.Exec(context.Background(), `
			DECLARE nft_income_migration_cursor CURSOR FOR
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
					incomeMap := utils.GetIncomeMap(events)
					for address, amount := range incomeMap {
						incomes = append(incomes, db.NftIncome{
							ClassId: classId,
							NftId:   nftId,
							TxHash:  txHash,
							Address: address,
							Amount:  amount,
						})
					}
				}
			}
			if pkeyId == 0 {
				break
			}
			count := len(incomes)
			if count == 0 {
				continue
			}
			for i := 0; i < count; i++ {
				income := incomes[i]
				_, err = dbTx.Exec(context.Background(), `
						INSERT INTO nft_income (class_id, nft_id, tx_hash, address, amount)
						VALUES ($1, $2, $3, $4, $5);
					`, income.ClassId, income.NftId, income.TxHash, income.Address, income.Amount)
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
			)
		}
		return nil
	})
}
