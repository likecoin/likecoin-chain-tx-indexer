package importdb

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgx/v4"
	"github.com/likecoin/likechain/app"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/state/txindex/kv"
	"github.com/tendermint/tendermint/store"
	dbm "github.com/tendermint/tm-db"
)

const batchSize = 10000

func Run(conn *pgx.Conn, likedPath string) {
	cdc = app.MakeCodec()
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

	lastHeight, err := db.GetLatestHeight(conn)
	startHeight := lastHeight
	if lastHeight == 0 {
		startHeight = 1
	}

	batch := db.NewBatch(conn, batchSize)
	for height := startHeight; height < maxHeight; height++ {
		block := blockStore.LoadBlock(height)
		txs := block.Data.Txs
		for txIndex, tx := range txs {
			txHash := cmn.HexBytes(tx.Hash())
			txResult, err := txIndexer.Get(txHash)
			var txRes sdk.TxResponse
			if err != nil || txResult == nil {
				logger.L.Warnw("Invalida transaction result, replacing with empty result", "txhash", txHash)
				txRes = sdk.TxResponse{Height: height, TxHash: txHash.String()}
			} else {
				txRes, err = formatTxResult(txHash, txResult, block)
				if err != nil {
					logger.L.Panicw("Cannot parse transaction", "txhash", txHash, "tx_raw", txResult.Tx, "error", err)
				}
			}
			txJSON, err := cdc.MarshalJSON(txRes)
			if err != nil {
				logger.L.Panicw("Cannot marshal tx response to JSON", "txhash", txHash, "tx_response", txRes, "error", err)
			}
			err = batch.InsertTx(txJSON, height, txIndex)
			if err != nil {
				logger.L.Panicw("Cannot insert transcation", "txhash", txHash, "tx_response", txRes, "tx_json", txJSON, "error", err)
			}
		}
	}
	err = batch.Flush()
	if err != nil {
		logger.L.Panicw("Cannot flush transcation batch", "batch", batch, "error", err)
	}
}