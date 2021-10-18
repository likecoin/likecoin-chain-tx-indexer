package db

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likechain/app"
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

const (
	ORDER_ASC  = "ASC"
	ORDER_DESC = "DESC"
)

var encodingConfig = app.MakeEncodingConfig()

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
	retryIntervals := []int{1, 1, 2, 5, 10, 15, 30, 30}
	maxRetry := len(retryIntervals)
	for i, retryInterval := range retryIntervals {
		ctx, cancel := GetTimeoutContext()
		defer cancel()
		pool, err := pgxpool.Connect(ctx, s)
		if err == nil || i == maxRetry-1 {
			return pool, err
		}
		logger.L.Errorw("Initialize connection pool failed, retrying", "error", err, "remaining_retry", maxRetry-i-1)
		time.Sleep(time.Duration(retryInterval) * time.Second)
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
	_, err := conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS txs (
		id BIGSERIAL PRIMARY KEY,
		height BIGINT,
		tx_index INT,
		tx JSONB,
		events VARCHAR ARRAY,
		UNIQUE (height, tx_index)
	)
	`)
	if err != nil {
		return err
	}
	ctx, cancel = GetTimeoutContext()
	defer cancel()
	_, err = conn.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_txs_txhash ON txs USING hash ((tx->>'txhash'))")
	if err != nil {
		return err
	}
	ctx, cancel = GetTimeoutContext()
	defer cancel()
	_, err = conn.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_txs_height_tx_index ON txs (height, tx_index)")
	if err != nil {
		return err
	}
	ctx, cancel = GetTimeoutContext()
	defer cancel()
	_, err = conn.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_tx_events ON txs USING gin (events)")
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

func getEventStrings(events types.StringEvents) []string {
	eventStrings := []string{}
	for _, event := range events {
		for _, attr := range event.Attributes {
			s := fmt.Sprintf("%s.%s=\"%s\"", event.Type, attr.Key, attr.Value)
			if len(s) < 8100 {
				// Cosmos SDK indeed generate meaninglessly long event strings
				// (e.g. in IBC client update, hex-encoding the whole header)
				// These event strings are useless and can't be handled by Postgres GIN index
				eventStrings = append(eventStrings, s)
			}
		}
	}
	return eventStrings
}

func QueryCount(conn *pgxpool.Conn, events types.StringEvents) (uint64, error) {
	sql := `
		SELECT count(*) FROM txs
		WHERE events @> $1
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	eventStrings := getEventStrings(events)
	row := conn.QueryRow(ctx, sql, eventStrings)
	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func QueryTxs(conn *pgxpool.Conn, events types.StringEvents, limit uint64, offset uint64, order string) ([]*types.TxResponse, error) {
	sql := fmt.Sprintf(`
		SELECT tx FROM txs
		WHERE events @> $1
		ORDER BY id %s
		LIMIT $2
		OFFSET $3
	`, order)
	eventStrings := getEventStrings(events)
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, eventStrings, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]*types.TxResponse, 0, limit)
	for rows.Next() {
		var jsonb pgtype.JSONB
		err := rows.Scan(&jsonb)
		if err != nil {
			return nil, fmt.Errorf("cannot scan rows to JSON: %+v", rows)
		}
		var txRes types.TxResponse
		err = encodingConfig.Marshaler.UnmarshalJSON(jsonb.Bytes, &txRes)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal JSON to TxResponse: %+v", jsonb.Bytes)
		}
		res = append(res, &txRes)
	}
	return res, nil
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

func (batch *Batch) InsertTx(txRes types.TxResponse, height int64, txIndex int) error {
	if batch.Batch.Len() >= batch.limit && batch.prevHeight > 0 && height != batch.prevHeight {
		err := batch.Flush()
		if err != nil {
			return err
		}
	}
	eventStrings := []string{}
	for _, log := range txRes.Logs {
		eventStrings = append(eventStrings, getEventStrings(log.Events)...)
	}
	txResJSON, err := encodingConfig.Marshaler.MarshalJSON(&txRes)
	if err != nil {
		return err
	}
	logger.L.Infow("Processing transaction", "txhash", txRes.TxHash, "height", height, "index", txIndex)
	batch.Batch.Queue("INSERT INTO txs (height, tx_index, tx, events) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", height, txIndex, txResJSON, eventStrings)
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
