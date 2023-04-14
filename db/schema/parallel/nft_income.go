package parallel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/rest"
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
		// use `WHERE rn = 1` to ensure only the latest event with same tx_hash is processed,
		// that way we can skip most unused `mint_nft` actions mixed in the `/cosmos.nft.v1beta1.MsgSend` actions
		_, err := dbTx.Exec(context.Background(), `
			DECLARE nft_income_migration_cursor CURSOR FOR
				SELECT id, class_id, nft_id, tx_hash, events
				FROM (
					SELECT e.id, e.class_id, e.nft_id, e.tx_hash, t.events, ROW_NUMBER() OVER (PARTITION BY e.tx_hash ORDER BY e.id DESC) AS rn
					FROM nft_event AS e
					JOIN (
						SELECT DISTINCT ON (tx ->> 'txhash') tx ->> 'txhash' AS tx_hash, events
						FROM txs
					) AS t ON e.tx_hash = t.tx_hash
					WHERE e.action IN ('/cosmos.nft.v1beta1.MsgSend', 'buy_nft', 'sell_nft')
				) subquery
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
				var eventRaw []string
				err = rows.Scan(&pkeyId, &classId, &nftId, &txHash, &eventRaw)
				if err != nil {
					logger.L.Errorw("Error when scanning row", "error", err)
					return err
				}

				// `spreadEvents` includes events of 'ALL' messages in the tx, not just a message.
				// this is different from the case in `extractNftIncomes()`
				spreadEvents, err := utils.ParseEvents(eventRaw)
				if err != nil {
					logger.L.Errorw("Error when parsing events", "error", err)
					return err
				}

				firstMsgAction := utils.GetEventsValue(spreadEvents, "message", "action")
				if firstMsgAction == "/cosmos.authz.v1beta1.MsgExec" || firstMsgAction == string(db.ACTION_BUY) || firstMsgAction == string(db.ACTION_SELL) {
					incomeMap := utils.GetIncomeMap(spreadEvents)
					for _, address := range rest.DefaultApiAddresses {
						delete(incomeMap, address)
					}
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
				logger.L.Infow(
					"NFT income table migration progress",
					"pkey_id", pkeyId,
					"count", count,
				)
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
		logger.L.Infow("NFT income table migration completed")
		return nil
	})
}
