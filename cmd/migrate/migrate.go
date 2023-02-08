package migrate

import (
	"github.com/spf13/cobra"
)

const (
	CmdBatchSize = "batch-size"
)

var MigrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Run application level migration in parallel with the poller",
}

func init() {
	MigrateCommand.AddCommand(
		MigrationSetupIscnVersionTableCommand,
		MigrationAddressPrefixCommand,
		MigrationNftEventPriceCommand,
		MigrationNftPriceCommand,
	)
}
