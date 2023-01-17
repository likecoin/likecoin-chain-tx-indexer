package migrate

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema/parallel"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func prepParams(cmd *cobra.Command) (conn *pgxpool.Conn, batchSize uint64, err error) {
	batchSize, err = cmd.Flags().GetUint64(CmdBatchSize)
	if err != nil {
		return nil, 0, err
	}
	pool, err := db.GetConnPoolFromCmdArgs(cmd)
	if err != nil {
		logger.L.Panicw("Cannot initialize database connection pool", "error", err)
	}
	conn, err = db.AcquireFromPool(pool)
	if err != nil {
		return nil, 0, err
	}
	return conn, batchSize, nil
}

var MigrationAddressPrefixCommand = &cobra.Command{
	Use:   "address-prefix",
	Short: "Migrate address prefix for iscn owner address, nft owner address, nft event sender and receiver addresses",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, batchSize, err := prepParams(cmd)
		if err != nil {
			return err
		}
		defer conn.Release()
		return parallel.MigrateAddressPrefix(conn, batchSize)
	},
}

var MigrateIscnOwnerCommand = &cobra.Command{
	Use:   "iscn-owner",
	Short: "Migrate address prefix for iscn owner address",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, batchSize, err := prepParams(cmd)
		if err != nil {
			return err
		}
		defer conn.Release()
		return parallel.MigrateIscnOwner(conn, batchSize)
	},
}

var MigrateNftOwner = &cobra.Command{
	Use:   "nft-owner",
	Short: "Migrate address prefix for NFT owner address",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, batchSize, err := prepParams(cmd)
		if err != nil {
			return err
		}
		defer conn.Release()
		return parallel.MigrateNftOwner(conn, batchSize)
	},
}

var MigrateNftEventSenderAndReceiver = &cobra.Command{
	Use:   "nft-event-sender-and-receiver",
	Short: "Migrate address prefix for NFT event sender and receiver address",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, batchSize, err := prepParams(cmd)
		if err != nil {
			return err
		}
		defer conn.Release()
		return parallel.MigrateNftEventSenderAndReceiver(conn, batchSize)
	},
}

func init() {
	MigrationAddressPrefixCommand.PersistentFlags().Uint64(
		CmdBatchSize,
		1000,
		"batch size when scanning tables for address prefixes",
	)
	MigrationAddressPrefixCommand.AddCommand(
		MigrateIscnOwnerCommand,
		MigrateNftOwner,
		MigrateNftEventSenderAndReceiver,
	)
}
