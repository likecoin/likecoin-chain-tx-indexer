package serve

import (
	"net/http"
	"time"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/extractor"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/poller"
	"github.com/likecoin/likecoin-chain-tx-indexer/rest"
	"github.com/likecoin/likecoin-chain/v3/app"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "serve",
	Short: "Run the indexing service and expose HTTP API",
	Run: func(cmd *cobra.Command, args []string) {
		pool, err := db.NewConnPoolFromCmdArgs(cmd)
		if err != nil {
			logger.L.Panicw("Cannot connect to Postgres", "error", err)
		}
		conn, err := db.AcquireFromPool(pool)
		if err != nil {
			logger.L.Panicw("Cannot acquire connection from database connection pool", "error", err)
		}
		err = db.InitDB(conn)
		if err != nil {
			logger.L.Panicw("Cannot initialize database", "error", err)
		}
		conn.Release()

		listenAddr, err := cmd.PersistentFlags().GetString("listen-addr")
		if err != nil {
			logger.L.Panicw("Cannot get listen address from command line parameters", "error", err)
		}
		lcdEndpoint, err := cmd.PersistentFlags().GetString("lcd-endpoint")
		if err != nil {
			logger.L.Panicw("Cannot get lcd endpoint address from command line parameters", "error", err)
		}
		if lcdEndpoint[len(lcdEndpoint)-1] == '/' {
			lcdEndpoint = lcdEndpoint[:len(lcdEndpoint)-1]
		}

		ctx := poller.CosmosCallContext{
			Codec: app.MakeEncodingConfig().Amino.Amino,
			Client: &http.Client{
				Transport: &http.Transport{
					MaxIdleConnsPerHost: 20,
				},
				Timeout: 10 * time.Second,
			},
			LcdEndpoint: lcdEndpoint,
		}

		go rest.Run(pool, listenAddr, lcdEndpoint)
		poller.Run(pool, &ctx, extractor.Run(pool))
	},
}

func init() {
	Command.PersistentFlags().String("lcd-endpoint", "http://localhost:1317", "LikeCoin chain lite client RPC endpoint")
	Command.PersistentFlags().String("listen-addr", "localhost:8997", "HTTP API serving address")
}
