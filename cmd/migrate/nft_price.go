package migrate

import (
	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema/parallel"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

var MigrationNftPriceCommand = &cobra.Command{
	Use:   "nft-price",
	Short: "Setup price column in nft table",
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
		return parallel.MigrateNftPrice(conn, batchSize)
	},
}

func init() {
	MigrationNftPriceCommand.PersistentFlags().Uint64(
		CmdBatchSize,
		1000,
		"number of ids in nft table to scan each time",
	)
}
