package poller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/tendermint/go-amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"
)

const batchSize = 1000

// TODO: move into config
const sleepInitial = 5 * time.Second
const sleepMax = 600 * time.Second

func getResponse(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 code returned: %d", resp.StatusCode)
	}
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

func getHeight(pool *pgxpool.Pool) (int64, error) {
	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	dbHeight, err := db.GetLatestHeight(conn)
	if err != nil {
		return 0, err
	}
	lastHeight := dbHeight - 1
	if lastHeight < 0 {
		lastHeight = 0
	}
	return lastHeight, nil
}

func poll(pool *pgxpool.Pool, ctx *CosmosCallContext, lastHeight int64) (int64, error) {
	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		return 0, fmt.Errorf("cannot acquire connection from database connection pool: %w", err)
	}
	defer conn.Release()
	batch := db.NewBatch(conn, batchSize)
	blockResult, err := GetBlock(ctx, 0)
	if err != nil {
		// TODO: retry
		return 0, fmt.Errorf("cannot get latest block from lcd: %w", err)
	}
	maxHeight := blockResult.Block.Height
	for height := lastHeight + 1; height <= maxHeight; height++ {
		blockResult, err := GetBlock(ctx, height)
		if err != nil {
			return 0, fmt.Errorf("cannot get block from lcd, error = %w, height = %d", err, height)
		}
		for txIndex, tx := range blockResult.Block.Data.Txs {
			txHash := cmn.HexBytes(tx.Hash())
			logger.L.Infow("Getting transaction", "txhash", txHash, "height", height, "index", txIndex)
			url := fmt.Sprintf("%s/txs/%s", ctx.LcdEndpoint, txHash.String())
			txResJSON, err := getResponse(ctx.Client, url)
			if err != nil {
				return 0, fmt.Errorf("cannot get tx response from lcd, error = %w, txhash = %s, height = %d, index = %d", err, txHash.String(), height, txIndex)
			}
			err = batch.InsertTx(txResJSON, height, txIndex)
			if err != nil {
				return 0, fmt.Errorf("cannot insert transaction, error = %w, txhash = %s, height = %d, index = %d, tx_json = %s", err, txHash.String(), height, txIndex, string(txResJSON))
			}
		}
	}
	err = batch.Flush()
	if err != nil {
		return 0, fmt.Errorf("cannot flush transaction batch, error = %w, batch = %v", err, batch)
	}
	return maxHeight, nil
}

func Run(pool *pgxpool.Pool, ctx *CosmosCallContext) {
	lastHeight, err := getHeight(pool)
	if err != nil {
		logger.L.Panicw("Cannot get height from database", "error", err)
	}
	toSleep := sleepInitial
	for {
		lastHeight, err = poll(pool, ctx, lastHeight)
		if err == nil {
			// reset sleep time to normal value
			toSleep = sleepInitial
		} else {
			logger.L.Errorw("cannot poll block", "error", err)
			// exponential back-off with max cap
			toSleep = toSleep * 2
			if toSleep > sleepMax {
				toSleep = sleepMax
			}
		}
		time.Sleep(toSleep)
	}
}
