package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/cobra"
)

const CmdDBName = "postgres-db"
const CmdDBHost = "postgres-host"
const CmdDBPort = "postgres-port"
const CmdDBUser = "postgres-user"
const CmdDBPassword = "postgres-pwd"

const DefaultDBName = "postgres"
const DefaultDBHost = "localhost"
const DefaultDBPort = "5432"
const DefaultDBUser = "postgres"
const DefaultDBPassword = "password"

func ConfigCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String(CmdDBName, DefaultDBName, "Postgres database name")
	cmd.PersistentFlags().String(CmdDBHost, DefaultDBHost, "Postgres host address")
	cmd.PersistentFlags().String(CmdDBPort, DefaultDBPort, "Postgres port")
	cmd.PersistentFlags().String(CmdDBUser, DefaultDBUser, "Postgres user")
	cmd.PersistentFlags().String(CmdDBPassword, DefaultDBPassword, "Postgres password")
}

func NewConnFromCmdArgs(cmd *cobra.Command) (*pgx.Conn, error) {
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
	s := fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s", dbname, host, port, user, pwd)
	return pgx.Connect(context.Background(), s)
}

func InitDB(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS txs (id BIGSERIAL PRIMARY KEY, height BIGINT, tx_index INT, tx JSONB, UNIQUE (height, tx_index))")
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), "CREATE INDEX IF NOT EXISTS idx_txs ON txs USING hash ((tx->>'txhash'))")
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), "CREATE INDEX IF NOT EXISTS idx_txs ON txs (height, tx_index)")
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS tx_events (type TEXT, key TEXT, value TEXT, height BIGINT, tx_index INT)")
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), "CREATE INDEX IF NOT EXISTS idx_tx_events ON tx_events (type, key, value, height, tx_index)")
	if err != nil {
		return err
	}
	return nil
}

func GetLatestHeight(conn *pgx.Conn) (int64, error) {
	rows, err := conn.Query(context.Background(), "SELECT max(height) FROM txs")
	if err != nil {
		fmt.Printf("'%s'\n", err.Error())
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
	Conn       *pgx.Conn
	Batch      pgx.Batch
	limit      int
	prevHeight int64
}

func NewBatch(conn *pgx.Conn, limit int) Batch {
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
	fmt.Printf("Inserting transaction %s (at height %d index %d)\n", txRes.TxHash, height, txIndex)
	batch.Batch.Queue("INSERT INTO txs (height, tx_index, tx) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", height, txIndex, txJSON)
	for _, log := range txRes.Logs {
		for _, event := range log.Events {
			for _, attr := range event.Attributes {
				batch.Batch.Queue(
					"INSERT INTO tx_events (type, key, value, height, tx_index) VALUES ($1, $2, $3, $4, $5)",
					event.Type, attr.Key, attr.Value, height, txIndex,
				)
			}
		}
	}
	batch.prevHeight = height
	fmt.Printf("Batch size = %d\n", batch.Batch.Len())
	return nil
}

func (batch *Batch) Flush() error {
	if batch.Batch.Len() > 0 {
		fmt.Println("Batch inserting into Postgres")
		result := batch.Conn.SendBatch(context.Background(), &batch.Batch)
		_, err := result.Exec()
		if err != nil {
			return err
		}
		result.Close()
		batch.Batch = pgx.Batch{}
	}
	return nil
}
