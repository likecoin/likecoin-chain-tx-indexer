package serve

import (
	"context"

	"github.com/likecoin/tm-postgres-indexer/db"
	"github.com/likecoin/tm-postgres-indexer/rest"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "serve",
	Short: "Run the indexing service and expose HTTP API",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := db.NewConnFromCmdArgs(cmd)
		if err != nil {
			panic(err)
		}
		defer conn.Close(context.Background())
		err = db.InitDB(conn)
		if err != nil {
			panic(err)
		}
		listenAddr, err := cmd.PersistentFlags().GetString("listen-addr")
		if err != nil {
			panic(err)
		}
		rest.Run(conn, listenAddr)
	},
}

func init() {
	Command.PersistentFlags().String("listen-addr", "localhost:8997", "HTTP API serving address")
}
