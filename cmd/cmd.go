package cmd

import (
	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain-tx-indexer/cmd/importdb"
	"github.com/likecoin/likecoin-chain-tx-indexer/cmd/migrate"
	"github.com/likecoin/likecoin-chain-tx-indexer/cmd/serve"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

var rootCmd = &cobra.Command{
	Use:   "indexer",
	Short: "The indexing service for LikeCoin chain transactions",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.SetupLoggerFromCmdArgs(cmd)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.L.Fatalw("Root command execution failed", "error", err)
	}
}

func init() {
	db.ConfigCmd(rootCmd)
	logger.ConfigCmd(rootCmd)
	rootCmd.AddCommand(
		importdb.Command,
		serve.Command,
		migrate.MigrateCommand,
	)
}
