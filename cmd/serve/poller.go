package serve

import (
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain/v3/app"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema"
	"github.com/likecoin/likecoin-chain-tx-indexer/extractor"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/poller"
	"github.com/likecoin/likecoin-chain-tx-indexer/pubsub"
)

var PollerCommand = &cobra.Command{
	Use:   "poller",
	Short: "Run the indexing service",
	Run: func(cmd *cobra.Command, args []string) {
		ServePoller(cmd)
	},
}

func ServePoller(cmd *cobra.Command) {
	pool, err := db.GetConnPoolFromCmdArgs(cmd)
	if err != nil {
		logger.L.Panicw("Cannot initialize database connection pool", "error", err)
	}

	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		logger.L.Panicw("Cannot acquire connection from database connection pool", "error", err)
	}
	defer conn.Release()
	err = schema.InitDB(conn)
	if err != nil {
		logger.L.Panicw("Cannot initialize database", "error", err)
	}
	conn.Release()

	lcdEndpoint, err := cmd.Flags().GetString("lcd-endpoint")
	if err != nil {
		logger.L.Panicw("Cannot get lcd endpoint address from command line parameters", "error", err)
	}

	err = pubsub.InitPubsubFromCmd(cmd)
	if err != nil {
		logger.L.Errorw("Pubsub initialization filed", "error", err)
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
	poller.Run(pool, &ctx, extractor.Run(pool))
}
