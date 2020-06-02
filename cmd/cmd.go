package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/likecoin/tm-postgres-indexer/cmd/importdb"
	"github.com/likecoin/tm-postgres-indexer/cmd/serve"
	"github.com/likecoin/tm-postgres-indexer/db"
)

var rootCmd = &cobra.Command{
	Use:   "indexer",
	Short: "The indexing service for LikeCoin chain transactions",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	db.ConfigCmd(rootCmd)
	rootCmd.AddCommand(importdb.Command)
	rootCmd.AddCommand(serve.Command)
}
