package poller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/tendermint/go-amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"
)

func getResponse(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type CosmosCallContext struct {
	Codec       *amino.Codec
	Client      *http.Client
	LcdEndpoint string
}

func GetBlock(ctx *CosmosCallContext, height int64) (*coreTypes.ResultBlock, error) {
	heightStr := "latest"
	if height > 0 {
		heightStr = fmt.Sprintf("%d", height)
	}
	url := fmt.Sprintf("%s/blocks/%s", ctx.LcdEndpoint, heightStr)
	body, err := getResponse(ctx.Client, url)
	if err != nil {
		return nil, err
	}
	resultBlock := coreTypes.ResultBlock{}
	err = ctx.Codec.UnmarshalJSON(body, &resultBlock)
	if err != nil {
		return nil, err
	}
	return &resultBlock, nil
}

type TxResult struct {
	Height int64 `json:"height"`
	Logs   []struct {
		Events []struct {
			Type       string `json:"type"`
			Attributes []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"attributes"`
		} `json:"events"`
	} `json:"logs"`
}

func Run(conn *pgx.Conn, ctx *CosmosCallContext) {
	batchSize := 1000
	batch := db.NewBatch(conn, batchSize)
	dbHeight, err := db.GetLatestHeight(conn)
	if err != nil {
		// TODO: retry
		logger.L.Panicw("Cannot get height from database", "error", err)
	}
	lastHeight := dbHeight - 1
	if lastHeight < 0 {
		lastHeight = 0
	}
	for {
		blockResult, err := GetBlock(ctx, 0)
		if err != nil {
			// TODO: retry
			logger.L.Panicw("Cannot get latest block from lcd", "error", err)
		}
		maxHeight := blockResult.Block.Height
		for height := lastHeight + 1; height <= maxHeight; height++ {
			blockResult, err := GetBlock(ctx, height)
			if err != nil {
				logger.L.Panicw("Cannot get block from lcd", "height", height, "error", err)
			}
			for txIndex, tx := range blockResult.Block.Data.Txs {
				txHash := cmn.HexBytes(tx.Hash())
				logger.L.Infow("Getting transaction", "txhash", txHash, "height", height, "index", txIndex)
				url := fmt.Sprintf("%s/txs/%s", ctx.LcdEndpoint, txHash.String())
				txResJSON, err := getResponse(ctx.Client, url)
				if err != nil {
					logger.L.Panicw("Cannot get tx response from lcd", "txhash", txHash, "height", height, "index", txIndex, "error", err)
				}
				err = batch.InsertTx(txResJSON, height, txIndex)
				if err != nil {
					logger.L.Panicw("Cannot insert transaction", "txhash", txHash, "height", height, "index", txIndex, "tx_json", txResJSON, "error", err)
				}
			}
		}
		err = batch.Flush()
		if err != nil {
			logger.L.Panicw("Cannot flush transaction batch", "batch", batch, "error", err)
		}
		lastHeight = maxHeight
		// TODO: move into config
		time.Sleep(5 * time.Second)
	}
}
