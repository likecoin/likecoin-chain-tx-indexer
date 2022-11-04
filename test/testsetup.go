package test

import (
	"context"
	"embed"
	"fmt"
	"log"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

var Pool *pgxpool.Pool

//go:embed test_data.sql
//go:embed test_cleanup.sql
var EmbedFS embed.FS

func SetupDbAndRunTest(m *testing.M, preRun func(pool *pgxpool.Pool)) {
	logger.SetupLogger(zapcore.DebugLevel, []string{"stdout"}, "console")
	var err error
	Pool, err = db.NewConnPool(
		utils.Env("DB_NAME", "postgres_test"),
		utils.Env("DB_HOST", "localhost"),
		utils.Env("DB_PORT", "5432"),
		utils.Env("DB_USER", "postgres"),
		utils.Env("DB_PASS", "password"),
		32,
		4,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer Pool.Close()

	// this connection is dedicated for the lock to make the schema checkings on another connection "happens after" the lock
	lockConn, err := db.AcquireFromPool(Pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer lockConn.Release()

	// explicitly locking the database, so database related tests will not run in parallel
	_, err = lockConn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS testing_lock ()")
	if err != nil {
		log.Fatalln(err)
	}

	tx, err := lockConn.Begin(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	_, err = tx.Exec(context.Background(), "LOCK testing_lock")
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback(context.Background())

	conn, err := db.AcquireFromPool(Pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()
	onProduction, err := checkTableExists(conn, "meta")
	testOnProduction := (utils.EnvInt("TEST_ON_PRODUCTION", 0) == 1)
	if err != nil {
		log.Fatalln(err)
	}
	if !onProduction {
		if testOnProduction {
			log.Println("WARNING: TEST_ON_PRODUCTION is set but now on testing database.")
		}
		err := db.InitDB(conn)
		if err != nil {
			log.Fatalln(err)
		}
		err = setupTestDatabase(conn)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		if testOnProduction {
			log.Println("WARNING: explicitly testing on production database.")
		} else {
			log.Fatalln("Database is not empty testing database, not testing on it unless TEST_ON_PRODUCTION=1 is set.")
		}
	}
	if preRun != nil {
		preRun(Pool)
	}
	m.Run()
	if !onProduction {
		err = cleanupTestDatabase(conn)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func checkTableExists(conn *pgxpool.Conn, tableName string) (bool, error) {
	row := conn.QueryRow(context.Background(), "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1", tableName)
	var exist int64
	err := row.Scan(&exist)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func RunEmbededSQLFile(conn *pgxpool.Conn, filename string) error {
	sqlBz, err := EmbedFS.ReadFile(filename)
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), string(sqlBz))
	return err
}

func setupTestDatabase(conn *pgxpool.Conn) error {
	return RunEmbededSQLFile(conn, "test_data.sql")
}

func cleanupTestDatabase(conn *pgxpool.Conn) error {
	return RunEmbededSQLFile(conn, "test_cleanup.sql")
}

func DebugSQL(conn *pgxpool.Conn, ctx context.Context, sql string, args ...interface{}) (err error) {
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
