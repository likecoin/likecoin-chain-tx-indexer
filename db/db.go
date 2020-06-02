package db

import (
	"context"
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
