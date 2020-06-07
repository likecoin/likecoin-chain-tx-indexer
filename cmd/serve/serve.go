package serve

import (
	"context"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/poller"
	"github.com/likecoin/likecoin-chain-tx-indexer/rest"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "serve",
	Short: "Run the indexing service and expose HTTP API",
	Run: func(cmd *cobra.Command, args []string) {
		restConn, err := db.NewConnFromCmdArgs(cmd)
		if err != nil {
			logger.L.Panicw("Cannot connect to Postgres", "error", err)
		}
		defer restConn.Close(context.Background())
		pollerConn, err := db.NewConnFromCmdArgs(cmd)
		if err != nil {
			logger.L.Panicw("Cannot connect to Postgres", "error", err)
		}
		defer pollerConn.Close(context.Background())
		err = db.InitDB(pollerConn)
		if err != nil {
			logger.L.Panicw("Cannot initialize Postgres database", "error", err)
		}
		listenAddr, err := cmd.PersistentFlags().GetString("listen-addr")
		if err != nil {
			logger.L.Panicw("Cannot get listen address from command line parameters", "error", err)
		}
		lcdEndpoint, err := cmd.PersistentFlags().GetString("lcd-endpoint")
		if err != nil {
			logger.L.Panicw("Cannot get lcd endpoint address from command line parameters", "error", err)
		}
		go rest.Run(restConn, listenAddr, lcdEndpoint)
		poller.Run(pollerConn, lcdEndpoint)
	},
}

func init() {
	Command.PersistentFlags().String("lcd-endpoint", "http://localhost:1317", "LikeCoin chain lite client RPC endpoint")
	Command.PersistentFlags().String("listen-addr", "localhost:8997", "HTTP API serving address")
}
