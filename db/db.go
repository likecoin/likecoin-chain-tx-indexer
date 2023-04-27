package db

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/pubsub"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
	"github.com/likecoin/likecoin-chain/v3/app"
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

const META_EXTRACTOR = "extractor_v1"
const META_BLOCK_HEIGHT = "latest_block_height"
const META_BLOCK_TIME_EPOCH_NS = "latest_block_time_epoch_ns"

var (
	pool     *pgxpool.Pool = nil
	poolLock               = &sync.Mutex{}
)

type Order string

const (
	ORDER_ASC  Order = "ASC"
	ORDER_DESC Order = "DESC"
)

var encodingConfig = app.MakeEncodingConfig()

var (
	MainAddressPrefix = "like"
	AddressPrefixes   = []string{MainAddressPrefix, "cosmos"}
)

func serializeTx(txRes *types.TxResponse) ([]byte, error) {
	txResJSON, err := encodingConfig.Marshaler.MarshalJSON(txRes)
	if err != nil {
		return nil, err
	}
	sanitizedJSON := utils.SanitizeJSON(txResJSON)
	return sanitizedJSON, nil
}

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
	return context.WithTimeout(context.Background(), 45*time.Second)
}

func GetConnPoolFromCmdArgs(cmd *cobra.Command) (*pgxpool.Pool, error) {
	poolLock.Lock()
	defer poolLock.Unlock()
	if pool == nil {
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
		returnPool, err := NewConnPool(dbname, host, port, user, pwd, poolMin, poolMax)
		if err != nil {
			return nil, err
		}
		pool = returnPool
	}
	return pool, nil
}

func NewConnPool(dbname string, host string, port string, user string, pwd string, poolMin int, poolMax int) (pool *pgxpool.Pool, err error) {
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

func GetLatestHeight(conn *pgxpool.Conn) (int64, error) {
	return GetMetaHeight(conn, META_BLOCK_HEIGHT)
}

func GetLatestBlockTime(conn *pgxpool.Conn) (time.Time, error) {
	ns, err := GetMetaHeight(conn, META_BLOCK_TIME_EPOCH_NS)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ns/1e9, ns%1e9).UTC(), nil
}

func QueryCount(conn *pgxpool.Conn, events types.StringEvents, height uint64) (uint64, error) {
	sql := `
		SELECT count(*) FROM txs
		WHERE events @> $1
		AND ($2 = 0 or height = $2)
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	eventStrings := utils.GetEventStrings(events)
	row := conn.QueryRow(ctx, sql, eventStrings, height)
	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func QueryTxs(conn *pgxpool.Conn, events types.StringEvents, height uint64, p PageRequest) ([]byte, []*types.TxResponse, error) {
	sql := fmt.Sprintf(`
		SELECT id, tx FROM txs
		WHERE events @> $1
		AND ($2 = 0 OR height = $2)
		AND ($3 = 0 OR id > $3)
		AND ($4 = 0 OR id < $4)
		ORDER BY id %s
		LIMIT $5
		OFFSET $6
	`, p.Order())
	eventStrings := utils.GetEventStrings(events)
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, eventStrings, height, p.After(), p.Before(), p.Limit, p.Offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	res := make([]*types.TxResponse, 0)
	id := uint64(0)
	for rows.Next() {
		var jsonb pgtype.JSONB
		err := rows.Scan(&id, &jsonb)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot scan rows to JSON: %+v", rows)
		}
		var txRes types.TxResponse
		err = encodingConfig.Marshaler.UnmarshalJSON(jsonb.Bytes, &txRes)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot unmarshal JSON to TxResponse: %+v", jsonb.Bytes)
		}
		res = append(res, &txRes)
	}
	var nextKey []byte
	if id > 0 {
		nextKey = make([]byte, 8)
		binary.LittleEndian.PutUint64(nextKey, id)
	}
	return nextKey, res, nil
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
		eventStrings = append(eventStrings, utils.GetEventStrings(log.Events)...)
	}
	txResJSON, err := serializeTx(&txRes)
	if err != nil {
		return err
	}
	_ = pubsub.Publish("NewTx", json.RawMessage(txResJSON))
	logger.L.Infow("Processing transaction", "txhash", txRes.TxHash, "height", height, "index", txIndex)
	batch.Batch.Queue("INSERT INTO txs (height, tx_index, tx, events) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", height, txIndex, txResJSON, eventStrings)
	batch.prevHeight = height
	logger.L.Debugw("Processed height", "height", height, "batch_size", batch.Batch.Len())
	return nil
}

func (batch *Batch) UpdateLatestBlockHeight(height int64) {
	batch.UpdateMetaHeight(META_BLOCK_HEIGHT, height)
}

func (batch *Batch) UpdateLatestBlockTime(timeStr string) error {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return err
	}
	batch.UpdateMetaHeight(META_BLOCK_TIME_EPOCH_NS, t.UTC().UnixNano())
	return nil
}

func (batch *Batch) Flush() error {
	if batch.Batch.Len() > 0 {
		logger.L.Debugw("Flushing Postgres batch", "batch_size", batch.Batch.Len())
		ctx, cancel := GetTimeoutContext()
		defer cancel()
		result := batch.Conn.SendBatch(ctx, &batch.Batch)
		_, err := result.Exec()
		if err != nil {
			logger.L.Debugw("Error when flushing Postgres batch", "err", err, "batch_size", batch.Batch.Len())
			return err
		}
		result.Close()
		batch.Batch = pgx.Batch{}
	}
	return nil
}
