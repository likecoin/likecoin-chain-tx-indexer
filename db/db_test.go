package db

import (
	"log"
	"testing"

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
