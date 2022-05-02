package db

import (
	"context"
	"log"
	"testing"

	"github.com/jackc/pgx/v4"
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

func debugSQL(tx pgx.Tx, ctx context.Context, sql string, args ...interface{}) (err error) {
	// add this line to debug SQL (only in test)
	// debugSQL(tx, ctx, sql, eventStrings, queryString, keywordString, pagination.getOffset(), pagination.Limit)
	rows, err := tx.Query(ctx, "EXPLAIN "+sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() && err == nil {
		var line string
		err = rows.Scan(&line)
		log.Println(line)
	}
	return err
}
