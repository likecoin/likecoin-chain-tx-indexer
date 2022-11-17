package migrate

import (
	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

const (
	CmdBatchSize = "batch-size"
)

var MigrationSetupIscnVersionTableCommand = &cobra.Command{
	Use:   "setup-iscn-version-table",
	Short: "Setup ISCN version table by scanning iscn table",
	RunE: func(cmd *cobra.Command, args []string) error {
		batchSize, err := cmd.Flags().GetUint64(CmdBatchSize)
		if err != nil {
			return err
		}
		pool, err := db.GetConnPoolFromCmdArgs(cmd)
		if err != nil {
			logger.L.Panicw("Cannot initialize database connection pool", "error", err)
		}
		conn, err := db.AcquireFromPool(pool)
		if err != nil {
			logger.L.Panicw("Cannot acquire connection from database connection pool", "error", err)
		}
		defer conn.Release()
		return schema.MigrateSetupIscnVersionTable(conn, batchSize)
	},
}

func init() {
	MigrationSetupIscnVersionTableCommand.PersistentFlags().Uint64(
		CmdBatchSize,
		1000,
		"batch size when scanning iscn table for setting up iscn version table",
	)
}
