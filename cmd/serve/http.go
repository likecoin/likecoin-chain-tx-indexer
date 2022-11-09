package serve

import (
	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/rest"
)

var HTTPCommand = &cobra.Command{
	Use:   "http",
	Short: "Expose HTTP API of the indexer",
	Run: func(cmd *cobra.Command, args []string) {
		ServeHTTP(cmd)
	},
}

func ServeHTTP(cmd *cobra.Command) {
	pool, err := db.GetConnPoolFromCmdArgs(cmd)
	if err != nil {
		logger.L.Panicw("Cannot initialize database connection pool", "error", err)
	}

	listenAddr, err := cmd.Flags().GetString("listen-addr")
	if err != nil {
		logger.L.Panicw("Cannot get listen address from command line parameters", "error", err)
	}
	lcdEndpoint, err := cmd.Flags().GetString("lcd-endpoint")
	if err != nil {
		logger.L.Panicw("Cannot get lcd endpoint address from command line parameters", "error", err)
	}
	if lcdEndpoint[len(lcdEndpoint)-1] == '/' {
		lcdEndpoint = lcdEndpoint[:len(lcdEndpoint)-1]
	}
	rest.Run(pool, listenAddr, lcdEndpoint)
}
