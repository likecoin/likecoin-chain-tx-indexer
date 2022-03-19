package db

import (
	"log"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"go.uber.org/zap/zapcore"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	logger.SetupLogger(zapcore.DebugLevel, []string{"stdout"}, "console")
	var err error
	pool, err = NewConnPool(
		"mydb",
		"localhost",
		"5432",
		"wancat",
		"password",
		32,
		4,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer pool.Close()
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	InitDB(conn)
	m.Run()
}

func TestQueryISCN(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	events := types.StringEvents{
		types.StringEvent{
			Type: "iscn_record",
			Attributes: []types.Attribute{
				{
					Key:   "iscn_id",
					Value: "iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
				},
			},
		},
	}
	QueryISCN(conn, events)
}
