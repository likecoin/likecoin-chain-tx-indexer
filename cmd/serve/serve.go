package serve

import (
	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain-tx-indexer/pubsub"
)

var Command = &cobra.Command{
	Use:   "serve",
	Short: "Run the indexing service and expose HTTP API",
	Long:  "Deprecated. Use the `rest` and `poll` subcommands to run HTTP API server and poller separately instead.",
	Run: func(cmd *cobra.Command, args []string) {
		go ServeHTTP(cmd)
		ServePoller(cmd)
	},
}

func init() {
	Command.PersistentFlags().String("lcd-endpoint", "http://localhost:1317", "LikeCoin chain lite client RPC endpoint")
	Command.PersistentFlags().String("listen-addr", "localhost:8997", "HTTP API serving address")
	Command.AddCommand(PollerCommand, HTTPCommand)
	pubsub.ConfigCmd(Command)
}
