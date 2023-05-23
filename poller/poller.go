package poller

import (
	"fmt"
	"io"
	"net/http"
	"time"

	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
	"github.com/likecoin/likecoin-chain/v4/app"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/bytes"

	tmTypes "github.com/tendermint/tendermint/types"
)

var batchSize = utils.EnvInt("BATCH_SIZE", 1000)
var batchMaxHeightDiff = int64(utils.EnvInt("BATCH_MAX_HEIGHT_DIFF", 1000))

// TODO: move into config
var sleepInitial = time.Duration(utils.EnvInt("SLEEP_INITIAL", 5)) * time.Second
var sleepMax = time.Duration(utils.EnvInt("SLEEP_MAX", 600)) * time.Second

var encodingConfig = app.MakeEncodingConfig()

func getResponse(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 code returned: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
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

type BlockResult struct {
	Block struct {
		Header struct {
			Height int64  `json:"height"`
			Time   string `json:"time"`
		} `json:"header"`
		Data struct {
			Txs tmTypes.Txs `json:"txs"`
		} `json:"data"`
	} `json:"block"`
}

func GetBlock(ctx *CosmosCallContext, height int64) (*BlockResult, error) {
	heightStr := "latest"
	if height > 0 {
		heightStr = fmt.Sprintf("%d", height)
	}
	url := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/%s", ctx.LcdEndpoint, heightStr)
	body, err := getResponse(ctx.Client, url)
	if err != nil {
		return nil, err
	}
	resultBlock := BlockResult{}
	err = ctx.Codec.UnmarshalJSON(body, &resultBlock)
	if err != nil {
		return nil, err
	}
	return &resultBlock, nil
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
	latestBlockResult, err := GetBlock(ctx, 0)
	if err != nil {
		// TODO: retry
		return 0, fmt.Errorf("cannot get latest block from lcd: %w", err)
	}
	maxHeight := latestBlockResult.Block.Header.Height
	if maxHeight-lastHeight > batchMaxHeightDiff {
		maxHeight = lastHeight + batchMaxHeightDiff
	}
	logger.L.Debugw("Querying blocks", "lastHeight", lastHeight, "maxHeight", maxHeight)
	for height := lastHeight + 1; height <= maxHeight; height++ {
		blockResult, err := GetBlock(ctx, height)
		if err != nil {
			return 0, fmt.Errorf("cannot get block from lcd, error = %w, height = %d", err, height)
		}
		for txIndex, tx := range blockResult.Block.Data.Txs {
			txHash := bytes.HexBytes(tx.Hash())
			logger.L.Infow("Getting transaction", "txhash", txHash, "height", height, "index", txIndex)
			url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", ctx.LcdEndpoint, txHash.String())
			txResJSON, err := getResponse(ctx.Client, url)
			if err != nil {
				return 0, fmt.Errorf("cannot get tx response from lcd, error = %w, txhash = %s, height = %d, index = %d", err, txHash.String(), height, txIndex)
			}
			txRes := txTypes.GetTxResponse{}
			err = encodingConfig.Marshaler.UnmarshalJSON(txResJSON, &txRes)
			if err != nil {
				return 0, fmt.Errorf("cannot unmarshal tx response to JSON, error = %w, txhash = %s, tx_response = %v", err, txHash, txResJSON)
			}
			err = batch.InsertTx(*txRes.TxResponse, height, txIndex)
			if err != nil {
				return 0, fmt.Errorf("cannot insert transaction, error = %w, txhash = %s, height = %d, index = %d, tx_json = %s", err, txHash.String(), height, txIndex, string(txResJSON))
			}
		}
	}
	batch.UpdateLatestBlockHeight(maxHeight)
	// error is ignored since fail to update block time is not critical
	_ = batch.UpdateLatestBlockTime(latestBlockResult.Block.Header.Time)
	err = batch.Flush()
	if err != nil {
		return 0, fmt.Errorf("cannot flush transaction batch, error = %w, batch = %v", err, batch)
	}
	return maxHeight, nil
}

func Run(pool *pgxpool.Pool, ctx *CosmosCallContext, triggers ...chan<- int64) {
	lastHeight, err := getHeight(pool)
	logger.L.Infow("Init Height", "lastHeight", lastHeight)
	if err != nil {
		logger.L.Panicw("Cannot get height from database", "error", err)
	}
	toSleep := sleepInitial
	for {
		returnedHeight, err := poll(pool, ctx, lastHeight)
		if err == nil {
			// reset sleep time to normal value
			toSleep = sleepInitial
			lastHeight = returnedHeight
			go func() {
				for _, trigger := range triggers {
					trigger <- returnedHeight
				}
			}()
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
