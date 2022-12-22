package test

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	sdk "github.com/cosmos/cosmos-sdk/types"
	iscntypes "github.com/likecoin/likecoin-chain/v3/x/iscn/types"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

const (
	ADDR_01_LIKE   = "like1qyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqewmlu9"
	ADDR_01_COSMOS = "cosmos1qyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq2j8al7"
	ADDR_02_LIKE   = "like1qgqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqm5jeaq"
	ADDR_02_COSMOS = "cosmos1qgqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqggwm7m"
	ADDR_03_LIKE   = "like1qvqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqz94m9r"
	ADDR_03_COSMOS = "cosmos1qvqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq3efexc"
	ADDR_04_LIKE   = "like1qsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqlfq4l2"
	ADDR_04_COSMOS = "cosmos1qsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqv4uhu3"
	ADDR_05_LIKE   = "like1q5qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqxc8h8f"
	ADDR_05_COSMOS = "cosmos1q5qqqqqqqqqqqqqqqqqqqqqqqqqqqqqq4ym4yj"
	ADDR_06_LIKE   = "like1qcqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqyzw3xv"
	ADDR_06_COSMOS = "cosmos1qcqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqh7jn9h"
	ADDR_07_LIKE   = "like1quqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqanfn70"
	ADDR_07_COSMOS = "cosmos1quqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqw043a5"
	ADDR_08_LIKE   = "like1pqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqql7w3tp"
	ADDR_08_COSMOS = "cosmos1pqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqvzjng6"
	ADDR_09_LIKE   = "like1pyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqx0fnnz"
	ADDR_09_COSMOS = "cosmos1pyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq4n43se"
	ADDR_10_LIKE   = "like1pgqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqy4q4j8"
	ADDR_10_COSMOS = "cosmos1pgqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqhfuh3u"
)

var Pool *pgxpool.Pool
var Conn *pgxpool.Conn

//go:embed test_cleanup_data.sql
//go:embed test_cleanup_table.sql
var EmbedFS embed.FS

func SetupDbAndRunTest(m *testing.M, preRun func(pool *pgxpool.Pool)) {
	// wrap into a func to respect `defer` while calling `os.Exit` when done
	code := func() int {
		logger.SetupLogger(zapcore.DebugLevel, []string{"stdout"}, "console")
		var err error
		Pool, err = db.NewConnPool(
			utils.Env("DB_NAME", "postgres_test"),
			utils.Env("DB_HOST", "localhost"),
			utils.Env("DB_PORT", "5433"),
			utils.Env("DB_USER", "postgres"),
			utils.Env("DB_PASS", "password"),
			32,
			4,
		)
		if err != nil {
			logger.L.Panic(err)
		}
		defer Pool.Close()

		Conn, err = db.AcquireFromPool(Pool)
		if err != nil {
			logger.L.Panic(err)
		}
		defer Conn.Release()
		err = setupTestDatabase(Conn)
		if err != nil {
			logger.L.Panic(err)
		}
		if preRun != nil {
			preRun(Pool)
		}
		code := m.Run()
		err = cleanupTestDatabase(Conn)
		if err != nil {
			logger.L.Panic(err)
		}
		return code
	}()
	os.Exit(code)
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

func setupTestDatabase(conn *pgxpool.Conn) (err error) {
	// explicitly locking the database, so database related tests will not run in parallel
	// some randomly generated number to make sure there is no conflict with other application
	_, err = conn.Exec(context.Background(), "SELECT pg_advisory_lock(4317656794910816416)")
	if err != nil {
		return err
	}
	onProduction, err := checkTableExists(conn, "meta")
	if err != nil {
		return err
	}
	if onProduction {
		return fmt.Errorf("database is not empty testing database")
	}
	err = schema.InitDB(conn)
	if err != nil {
		return err
	}
	return nil
}

func cleanupTestDatabase(conn *pgxpool.Conn) error {
	return RunEmbededSQLFile(conn, "test_cleanup_table.sql")
}

func CleanupTestData(conn *pgxpool.Conn) error {
	return RunEmbededSQLFile(conn, "test_cleanup_data.sql")
}

type DBTestData struct {
	Iscns      []db.IscnInsert
	NftClasses []db.NftClass
	Nfts       []db.Nft
	NftEvents  []db.NftEvent
	Txs        []string
}

func InsertTestData(testData DBTestData) error {
	b := db.NewBatch(Conn, 10000)
	for _, i := range testData.Iscns {
		iscnId, err := iscntypes.ParseIscnId(i.Iscn)
		if err != nil {
			return err
		}
		i.IscnPrefix = iscnId.Prefix.String()
		i.Version = int(iscnId.Version)
		if i.Data == nil {
			i.Data = []byte("{}")
		}
		i.Timestamp = i.Timestamp.UTC()
		b.InsertIscn(i)
	}
	for _, c := range testData.NftClasses {
		c.Parent.Type = "ISCN"
		c.CreatedAt = c.CreatedAt.UTC()
		b.InsertNftClass(c)
	}
	for _, n := range testData.Nfts {
		b.InsertNft(n)
	}
	for _, e := range testData.NftEvents {
		e.Timestamp = e.Timestamp.UTC()
		b.InsertNftEvent(e)
	}
	for i, tx := range testData.Txs {
		height := 1
		type Log struct {
			Events sdk.StringEvents `json:"events,omitempty"`
		}
		logs := []Log{}
		txStruct := struct {
			Height string `json:"height,omitempty"`
			Logs   []Log  `json:"logs,omitempty"`
		}{}
		err := json.Unmarshal([]byte(tx), &txStruct)
		if err == nil {
			logs = txStruct.Logs
			h, err := strconv.Atoi(txStruct.Height)
			if err == nil && h > 0 {
				height = h
			}
		}

		eventStrings := []string{}
		for _, log := range logs {
			eventStrings = append(eventStrings, utils.GetEventStrings(log.Events)...)
		}
		b.Batch.Queue("INSERT INTO txs (height, tx_index, tx, events) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", height, i, []byte(tx), eventStrings)
		b.Batch.Queue("UPDATE meta SET height = $1 WHERE id = $2 AND height < $1", height, db.META_BLOCK_HEIGHT)
	}
	return b.Flush()
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
