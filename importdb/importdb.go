package importdb

import (
	"fmt"

	dbm "github.com/cometbft/cometbft-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/state/txindex/kv"
	"github.com/tendermint/tendermint/store"
)

const batchSize = 10000

func Run(pool *pgxpool.Pool, likedPath string) {
	likedDataDir := fmt.Sprintf("%s/data", likedPath)
	blockDB, err := dbm.NewGoLevelDB("blockstore", likedDataDir)
	if err != nil {
		logger.L.Panicw("Cannot initialize blockstore database from liked data", "error", err)
	}
	defer blockDB.Close()
	blockStore := store.NewBlockStore(blockDB)
	txIndexDB, err := dbm.NewGoLevelDB("tx_index", likedDataDir)
	if err != nil {
		logger.L.Panicw("Cannot initialize tx_index database from liked data", "error", err)
	}
	defer txIndexDB.Close()
	txIndexer := kv.NewTxIndex(txIndexDB)
	maxHeight := blockStore.Height()

	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		logger.L.Panicw("Cannot acquire connection from database connection pool", "error", err)
	}
	defer conn.Release()
	lastHeight, err := db.GetLatestHeight(conn)
	if err != nil {
		logger.L.Panicw("Cannot get height from database", "error", err)
	}
	startHeight := lastHeight
	if lastHeight <= 0 {
		startHeight = 1
	}

	batch := db.NewBatch(conn, batchSize)
	for height := startHeight; height < maxHeight; height++ {
		block := blockStore.LoadBlock(height)
		txs := block.Data.Txs
		for txIndex, tx := range txs {
			txHash := bytes.HexBytes(tx.Hash())
			txResult, err := txIndexer.Get(txHash)
			var txRes sdk.TxResponse
			if err != nil || txResult == nil {
				logger.L.Warnw("Invalid transaction result, replacing with empty result", "txhash", txHash)
				txRes = sdk.TxResponse{Height: height, TxHash: txHash.String()}
			} else {
				txRes, err = formatTxResult(txHash, txResult, block)
				if err != nil {
					logger.L.Panicw("Cannot parse transaction", "txhash", txHash, "tx_raw", txResult.Tx, "error", err)
				}
			}
			logger.L.Debugw("before Marshal JSON", "tx", txRes.Tx)
			err = batch.InsertTx(txRes, height, txIndex)
			if err != nil {
				logger.L.Panicw("Cannot insert transcation", "txhash", txHash, "tx_response", txRes, "error", err)
			}
		}
	}
	err = batch.Flush()
	if err != nil {
		logger.L.Panicw("Cannot flush transcation batch", "batch", batch, "error", err)
	}
}
