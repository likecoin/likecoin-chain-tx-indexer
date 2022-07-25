package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
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

type Order string

const (
	ORDER_ASC  Order = "ASC"
	ORDER_DESC Order = "DESC"
)

var encodingConfig = app.MakeEncodingConfig()

func serializeTx(txRes *types.TxResponse) ([]byte, error) {
	txResJSON, err := encodingConfig.Marshaler.MarshalJSON(txRes)
	if err != nil {
		return nil, err
	}
	// "\u0000" is not processible by Postgres
	// Below is a hotfix, long term solution is to store binary or migrate to Tendermint Postgres indexing backend
	sanitizedJSON := strings.ReplaceAll(string(txResJSON), "\\u0000", "")
	return []byte(sanitizedJSON), nil
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
	return NewConnPool(dbname, host, port, user, pwd, poolMin, poolMax)
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
	_, err = conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS iscn (
		id bigserial primary key,
		iscn_id text,
		iscn_id_prefix text,
		version int,
		owner text,
		name text,
		description text,
		url text,
		keywords text[],
		fingerprints text[],
		ipld text,
		timestamp timestamp,
		stakeholders jsonb,
		data jsonb,
		UNIQUE(iscn_id)
	)`)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS meta (
		id text PRIMARY KEY,
		height BIGINT
	)`)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, `
	INSERT INTO meta VALUES ($1, 0)
	ON CONFLICT DO NOTHING
	`, META_EXTRACTOR)
	if err != nil {
		return err
	}
	// ignore error of type alread exists
	conn.Exec(ctx, `
	CREATE TYPE class_parent_type AS ENUM ('UNKNOWN', 'ISCN', 'ACCOUNT')
	`)

	_, err = conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS nft_class (
		id bigserial primary key,
		class_id text unique,
		parent_type class_parent_type,
		parent_iscn_id_prefix text,
		parent_account text,
		name text,
		symbol text,
		description text,
		uri text,
		uri_hash text,
		metadata jsonb,
		config jsonb,
		price int
	);`)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS nft (
		id bigserial primary key,
		class_id text,
		owner text,
		nft_id text,
		uri text,
		uri_hash text,
		metadata jsonb,
		UNIQUE(class_id, nft_id)
	);`)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS nft_event (
		id bigserial primary key,
		action text,
		class_id text,
		nft_id text,
		sender text,
		receiver text,
		events text[],
		tx_hash text,
		timestamp timestamp,
		UNIQUE(action, class_id, nft_id, tx_hash)
	);`)
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
	// TODO: add support to not only 0 in the future
	ctx = context.Background()
	_, err = conn.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_record ON txs USING GIN((tx #> '{tx, body, messages, 0, record}') jsonb_path_ops)`)
	if err != nil {
		return err
	}
	if _, err = conn.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_keywords ON txs USING GIN((string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',')))`); err != nil {
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

func QueryCount(conn *pgxpool.Conn, events types.StringEvents, height uint64) (uint64, error) {
	sql := `
		SELECT count(*) FROM txs
		WHERE events @> $1
		AND ($2 = 0 or height = $2)
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	eventStrings := getEventStrings(events)
	row := conn.QueryRow(ctx, sql, eventStrings, height)
	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func QueryTxs(conn *pgxpool.Conn, events types.StringEvents, height uint64, limit uint64, offset uint64, order Order) ([]*types.TxResponse, error) {
	sql := fmt.Sprintf(`
		SELECT tx FROM txs
		WHERE events @> $1
		AND ($2 = 0 OR height = $2)
		ORDER BY id %s
		LIMIT $3
		OFFSET $4
	`, order)
	eventStrings := getEventStrings(events)
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, eventStrings, height, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseRows(rows, limit)
}

func parseRows(rows pgx.Rows, limit uint64) ([]*types.TxResponse, error) {
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
	txResJSON, err := serializeTx(&txRes)
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
