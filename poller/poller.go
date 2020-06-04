package poller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/likecoin/likechain/app"
	"github.com/likecoin/tm-postgres-indexer/db"
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

func getBlock(ctx *CosmosCallContext, height int64) (*coreTypes.ResultBlock, error) {
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

func Run(conn *pgx.Conn, lcdEndpoint string) {
	if lcdEndpoint[len(lcdEndpoint)-1] == '/' {
		lcdEndpoint = lcdEndpoint[:len(lcdEndpoint)-1]
	}
	ctx := CosmosCallContext{
		Codec:       app.MakeCodec(),
		Client:      &http.Client{},
		LcdEndpoint: lcdEndpoint,
	}

	batchSize := 1000
	batch := db.NewBatch(conn, batchSize)
	dbHeight, err := db.GetLatestHeight(conn)
	if err != nil {
		panic(err)
	}
	lastHeight := dbHeight - 1
	if lastHeight < 0 {
		lastHeight = 0
	}
	for {
		blockResult, err := getBlock(&ctx, 0)
		if err != nil {
			panic(err)
		}
		maxHeight := blockResult.Block.Height
		for height := lastHeight + 1; height <= maxHeight; height++ {
			blockResult, err := getBlock(&ctx, height)
			if err != nil {
				panic(err)
			}
			for txIndex, tx := range blockResult.Block.Data.Txs {
				txHash := cmn.HexBytes(tx.Hash())
				fmt.Printf("Getting tx %s (%d, %d)\n", txHash, height, txIndex)
				url := fmt.Sprintf("%s/txs/%s", ctx.LcdEndpoint, txHash.String())
				txResJSON, err := getResponse(ctx.Client, url)
				if err != nil {
					panic(err)
				}
				err = batch.InsertTx(txResJSON, height, txIndex)
				if err != nil {
					panic(err)
				}
			}
		}
		err = batch.Flush()
		if err != nil {
			panic(err)
		}
		lastHeight = maxHeight
		time.Sleep(5 * time.Second)
	}
}
