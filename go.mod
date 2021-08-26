module github.com/likecoin/likecoin-chain-tx-indexer

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.42.7
	github.com/gin-gonic/gin v1.7.4
	github.com/jackc/pgx/v4 v4.13.0
	github.com/likecoin/likechain v0.0.0-20210714072515-32056fc4759d
	github.com/spf13/cobra v1.1.3
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.11
	github.com/tendermint/tm-db v0.6.4
	go.uber.org/zap v1.13.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
