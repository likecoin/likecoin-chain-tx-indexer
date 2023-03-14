package migrate

import (
	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema/parallel"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

var MigrationNftRoyaltyCommand = &cobra.Command{
	Use:   "nft-royalty",
	Short: "Setup nft_royalty table by scanning nft_event table",
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
		return parallel.MigrateNftRoyalty(conn, batchSize)
	},
}

func init() {
	MigrationNftRoyaltyCommand.PersistentFlags().Uint64(
		CmdBatchSize,
		1000,
		"number of ids in nft_event table to scan each time",
	)
}
