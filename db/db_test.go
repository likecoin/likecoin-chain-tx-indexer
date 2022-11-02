package db

import (
	"context"
	"embed"
	"fmt"
	"log"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	. "github.com/likecoin/likecoin-chain-tx-indexer/utils"
	"go.uber.org/zap/zapcore"
)

var pool *pgxpool.Pool
var conn *pgxpool.Conn

//go:embed test_data.sql
//go:embed test_cleanup.sql
var embedFS embed.FS

func TestMain(m *testing.M) {
	godotenv.Load("../.env")
	logger.SetupLogger(zapcore.DebugLevel, []string{"stdout"}, "console")
	var err error
	pool, err = NewConnPool(
		Env("DB_NAME", "postgres_test"),
		Env("DB_HOST", "localhost"),
		Env("DB_PORT", "5432"),
		Env("DB_USER", "postgres"),
		Env("DB_PASS", "password"),
		32,
		4,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer pool.Close()
	conn, err = AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	err = InitDB(conn)
	if err != nil {
		log.Fatalln(err)
	}
	originalHeight, err := GetLatestHeight(conn)
	if err != nil {
		log.Fatalln(err)
	}
	if originalHeight == 0 {
		err := setupTestDatabase(conn)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		if EnvInt("TEST_ON_PRODUCTION", 0) == 1 {
			log.Println("WARNING: explicitly testing on production server.")
		} else {
			log.Fatalln("Database is not empty testing database, not testing on it unless TEST_ON_PRODUCTION=1 is set.")
		}
	}
	m.Run()
	if originalHeight == 0 {
		err := cleanupTestDatabase(conn)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func runEmbededSQLFile(conn *pgxpool.Conn, filename string) error {
	sqlBz, err := embedFS.ReadFile(filename)
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), string(sqlBz))
	return err
}

func setupTestDatabase(conn *pgxpool.Conn) error {
	return runEmbededSQLFile(conn, "test_data.sql")
}

func cleanupTestDatabase(conn *pgxpool.Conn) error {
	return runEmbededSQLFile(conn, "test_cleanup.sql")
}

func debugSQL(conn *pgxpool.Conn, ctx context.Context, sql string, args ...interface{}) (err error) {
	// add this line to debug SQL (only in test)
	// debugSQL(tx, ctx, sql, eventStrings, queryString, keywordString, pagination.getOffset(), pagination.Limit)
	rows, err := conn.Query(ctx, "EXPLAIN "+sql, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() && err == nil {
		var line string
		err = rows.Scan(&line)
		fmt.Println(line)
	}
	return err
}
