package importdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/likecoin/likechain/app"
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
		panic(err)
	}
	defer blockDB.Close()
	blockStore := store.NewBlockStore(blockDB)
	txIndexDB, err := dbm.NewGoLevelDB("tx_index", likedDataDir)
	if err != nil {
		panic(err)
	}
	defer txIndexDB.Close()
	txIndexer := kv.NewTxIndex(txIndexDB)
	maxHeight := blockStore.Height()
	batch := &pgx.Batch{}
	// TODO: select max height from Postgres
	for height := int64(1); height < maxHeight; height++ {
		if height%10000 == 0 {
			fmt.Printf("Processing block height %d\n", height)
		}
		block := blockStore.LoadBlock(height)
		txs := block.Data.Txs
		for txIndex, tx := range txs {
			txHash := cmn.HexBytes(tx.Hash())
			fmt.Printf("Importing transaction %s\n", txHash)
			txResult, err := txIndexer.Get(txHash)
			var txRes sdk.TxResponse
			if err != nil || txResult == nil {
				fmt.Printf("Warning: invalid result for transaction %s, replacing with empty txResult\n", txHash)
				txRes = sdk.TxResponse{Height: height, TxHash: txHash.String()}
			} else {
				txRes, err = formatTxResult(txHash, txResult, block)
				if err != nil {
					panic(err)
				}
			}
			txJSON, err := cdc.MarshalJSON(txRes)
			if err != nil {
				panic(err)
			}
			batch.Queue("INSERT INTO txs (height, tx_index, tx) VALUES ($1, $2, $3)", height, txIndex, txJSON)
			for _, log := range txRes.Logs {
				for _, event := range log.Events {
					for _, attr := range event.Attributes {
						batch.Queue(
							"INSERT INTO tx_events (type, key, value, height, tx_index) VALUES ($1, $2, $3, $4, $5)",
							event.Type, attr.Key, attr.Value, height, txIndex,
						)
					}
				}
			}
		}
		if batch.Len() >= batchSize || height >= maxHeight-1 {
			fmt.Printf("Batch inserting into Postgres")
			result := conn.SendBatch(context.Background(), batch)
			_, err := result.Exec()
			if err != nil {
				panic(err)
			}
			result.Close()
			batch = &pgx.Batch{}
		}
	}
}
