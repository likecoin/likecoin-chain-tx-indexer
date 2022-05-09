module github.com/likecoin/likecoin-chain-tx-indexer

go 1.16

require (
	github.com/armon/go-metrics v0.3.11 // indirect
	github.com/cosmos/cosmos-sdk v0.44.8
	github.com/gin-gonic/gin v1.7.4
	github.com/jackc/pgtype v1.8.1
	github.com/jackc/pgx/v4 v4.13.0
	github.com/likecoin/likechain v1.2.1-0.20220428063414-b79a2e611c71
	github.com/spf13/cobra v1.4.0
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.19
	github.com/tendermint/tm-db v0.6.6
	go.uber.org/zap v1.19.1
)

// point sdk to fork and follow replaces at https://github.com/cosmos/cosmos-sdk/blob/v0.44.8/go.mod
replace (
	github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
	github.com/cosmos/cosmos-sdk => github.com/likecoin/cosmos-sdk v0.44.8-dual-prefix
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
