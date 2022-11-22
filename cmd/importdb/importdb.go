package importdb

import (
	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema"
	"github.com/likecoin/likecoin-chain-tx-indexer/importdb"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

var Command = &cobra.Command{
	Use:   "import",
	Short: "Import from existing LikeCoin chain database",
	Run: func(cmd *cobra.Command, args []string) {
		likedPath, err := cmd.PersistentFlags().GetString("liked-path")
		if err != nil {
			logger.L.Panicw("Cannot get liked data folder path from command line parameters", "error", err)
		}
		pool, err := db.GetConnPoolFromCmdArgs(cmd)
		if err != nil {
			logger.L.Panicw("Cannot connect to Postgres", "error", err)
		}
		conn, err := db.AcquireFromPool(pool)
		if err != nil {
			logger.L.Panicw("Cannot acquire connection from database connection pool", "error", err)
		}
		err = schema.InitDB(conn)
		if err != nil {
			logger.L.Panicw("Cannot initialize Postgres database", "error", err)
		}
		conn.Release()
		importdb.Run(pool, likedPath)
	},
}

func Execute() {
	if err := Command.Execute(); err != nil {
		logger.L.Fatalw("Import command execution failed", "error", err)
	}
}

func init() {
	Command.PersistentFlags().String("liked-path", "./.liked", "location of the LikeCoin chain database folder (.liked)")
}
