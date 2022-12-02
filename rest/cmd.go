package rest

import (
	"github.com/spf13/cobra"
)

const (
	CmdLcdEndpoint  = "lcd-endpoint"
	CmdListenAddr   = "listen-addr"
	CmdApiAddresses = "api-address"

	DefaultLcdEndpoint = "http://localhost:1317"
	DefaultListenAddr  = "localhost:8997"
)

var (
	DefaultApiAddresses = []string{"like17m4vwrnhjmd20uu7tst7nv0kap6ee7js69jfrs"}
)

func ConfigCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String(CmdLcdEndpoint, DefaultLcdEndpoint, "LikeCoin chain lite client RPC endpoint")
	cmd.PersistentFlags().String(CmdListenAddr, DefaultListenAddr, "HTTP API serving address")
	cmd.PersistentFlags().StringSlice(CmdApiAddresses, DefaultApiAddresses, "Default API sender addresses for NFT ranking and stats")
}
