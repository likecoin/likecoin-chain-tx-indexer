package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/spf13/cobra"
)

const CmdDBName = "postgres-db"
const CmdDBHost = "postgres-host"
const CmdDBPort = "postgres-port"
const CmdDBUser = "postgres-user"
const CmdDBPassword = "postgres-pwd"
const CmdDBPoolMin = "postgres-pool-min"
const CmdDBPoolMax = "postgres-pool-max"

const DefaultDBName = "postgres"
const DefaultDBHost = "localhost"
const DefaultDBPort = "5432"
const DefaultDBUser = "postgres"
const DefaultDBPassword = "password"
const DefaultDBPoolMin = 4
const DefaultDBPoolMax = 32

func ConfigCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String(CmdDBName, DefaultDBName, "Postgres database name")
	cmd.PersistentFlags().String(CmdDBHost, DefaultDBHost, "Postgres host address")
	cmd.PersistentFlags().String(CmdDBPort, DefaultDBPort, "Postgres port")
	cmd.PersistentFlags().String(CmdDBUser, DefaultDBUser, "Postgres user")
	cmd.PersistentFlags().String(CmdDBPassword, DefaultDBPassword, "Postgres password")
	cmd.PersistentFlags().Int(CmdDBPoolMin, DefaultDBPoolMin, "Postgres minimum number of connections in connection pool")
	cmd.PersistentFlags().Int(CmdDBPoolMax, DefaultDBPoolMax, "Postgres maximum number of connections in connection pool")
}

func GetTimeoutContext() (context.Context, context.CancelFunc) {
	// TODO: move into config
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func NewConnPoolFromCmdArgs(cmd *cobra.Command) (pool *pgxpool.Pool, err error) {
	dbname, err := cmd.Flags().GetString(CmdDBName)
	if err != nil {
		return nil, err
	}
	host, err := cmd.Flags().GetString(CmdDBHost)
	if err != nil {
		return nil, err
	}
	port, err := cmd.Flags().GetString(CmdDBPort)
	if err != nil {
		return nil, err
	}
	user, err := cmd.Flags().GetString(CmdDBUser)
	if err != nil {
		return nil, err
	}
	pwd, err := cmd.Flags().GetString(CmdDBPassword)
	if err != nil {
		return nil, err
	}
	poolMin, err := cmd.Flags().GetInt(CmdDBPoolMin)
	if err != nil {
		return nil, err
	}
	poolMax, err := cmd.Flags().GetInt(CmdDBPoolMax)
	if err != nil {
		return nil, err
	}
	s := fmt.Sprintf(
		"dbname=%s host=%s port=%s user=%s password=%s pool_min_conns=%d pool_max_conns=%d",
		dbname, host, port, user, pwd, poolMin, poolMax,
	)
	maxRetry := 5
	for i := 0; i < maxRetry; i++ {
		ctx, cancel := GetTimeoutContext()
		defer cancel()
		pool, err := pgxpool.Connect(ctx, s)
		if err == nil || i == maxRetry-1 {
			return pool, err
		}
		logger.L.Errorw("Initialize connection pool failed, retrying", "error", err, "remaining_retry", 4-i)
		time.Sleep(time.Duration(1<<i) * time.Second)
	}
	return nil, err
}

func AcquireFromPool(pool *pgxpool.Pool) (*pgxpool.Conn, error) {
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	return pool.Acquire(ctx)
}

func InitDB(conn *pgxpool.Conn) error {
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	_, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS txs (id BIGSERIAL PRIMARY KEY, height BIGINT, tx_index INT, tx JSONB, UNIQUE (height, tx_index))")
	if err != nil {
		return err
	}
	ctx, cancel = GetTimeoutContext()
	defer cancel()
	_, err = conn.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_txs ON txs USING hash ((tx->>'txhash'))")
	if err != nil {
		return err
	}
	ctx, cancel = GetTimeoutContext()
	defer cancel()
	_, err = conn.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_txs ON txs (height, tx_index)")
	if err != nil {
		return err
	}
	ctx, cancel = GetTimeoutContext()
	defer cancel()
	_, err = conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS tx_events (type TEXT, key TEXT, value TEXT, height BIGINT, tx_index INT, UNIQUE(type, key, value, height, tx_index))")
	if err != nil {
		return err
	}
	ctx, cancel = GetTimeoutContext()
	defer cancel()
	_, err = conn.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_tx_events ON tx_events (type, key, value, height, tx_index)")
	if err != nil {
		return err
	}
	return nil
}

func GetLatestHeight(conn *pgxpool.Conn) (int64, error) {
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, "SELECT max(height) FROM txs")
	if err != nil {
		logger.L.Warnw("Cannot get latest height from Postgres", "error", err)
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		// No record in database, so max height = 0
		return 0, nil
	}
	var height int64
	err = rows.Scan(&height)
	if err != nil {
		return 0, nil
	}
	return height, nil
}

type Batch struct {
	Conn       *pgxpool.Conn
	Batch      pgx.Batch
	limit      int
	prevHeight int64
}

func NewBatch(conn *pgxpool.Conn, limit int) Batch {
	return Batch{
		Conn:       conn,
		Batch:      pgx.Batch{},
		limit:      limit,
		prevHeight: 0,
	}
}

type TxResult struct {
	TxHash string `json:"txhash"`
	Logs   []struct {
		Events []struct {
			Type       string `json:"type"`
			Attributes []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"attributes"`
		} `json:"events"`
	} `json:"logs"`
}

func (batch *Batch) InsertTx(txJSON []byte, height int64, txIndex int) error {
	if batch.Batch.Len() >= batch.limit && batch.prevHeight > 0 && height != batch.prevHeight {
		err := batch.Flush()
		if err != nil {
			return err
		}
	}
	txRes := TxResult{}
	err := json.Unmarshal(txJSON, &txRes)
	if err != nil {
		return err
	}
	logger.L.Infow("Processing transaction", "txhash", txRes.TxHash, "height", height, "index", txIndex)
	batch.Batch.Queue("INSERT INTO txs (height, tx_index, tx) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", height, txIndex, txJSON)
	for _, log := range txRes.Logs {
		for _, event := range log.Events {
			for _, attr := range event.Attributes {
				batch.Batch.Queue(
					"INSERT INTO tx_events (type, key, value, height, tx_index) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING",
					event.Type, attr.Key, attr.Value, height, txIndex,
				)
			}
		}
	}
	batch.prevHeight = height
	logger.L.Debugw("Processed height", "height", height, "batch_size", batch.Batch.Len())
	return nil
}

func (batch *Batch) Flush() error {
	if batch.Batch.Len() > 0 {
		logger.L.Debugw("Inserting transactions into Postgres in batch", "batch_size", batch.Batch.Len())
		ctx, cancel := GetTimeoutContext()
		defer cancel()
		result := batch.Conn.SendBatch(ctx, &batch.Batch)
		_, err := result.Exec()
		if err != nil {
			return err
		}
		result.Close()
		batch.Batch = pgx.Batch{}
	}
	return nil
}
