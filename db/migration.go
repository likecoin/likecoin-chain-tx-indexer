package db

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func GetSchemaVersion(dbTx pgx.Tx) (uint64, error) {
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	row := dbTx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT FROM information_schema.tables
  WHERE table_name = 'meta'
)`)
	var tableExist bool
	err := row.Scan(&tableExist)
	if err != nil {
		logger.L.Errorw("Error when querying whether 'meta' table exists", "error", err)
		return 0, err
	}
	if !tableExist {
		// table not exist, meaning that schema version = 0
		logger.L.Debug("Table meta does not exist, schema version = 0")
		return 0, nil
	}
	row = dbTx.QueryRow(ctx, `SELECT height FROM meta WHERE id = 'schema_version'`)
	var schemaVersion uint64
	err = row.Scan(&schemaVersion)
	if err != nil {
		if err == pgx.ErrNoRows {
			// table exist but no 'schema_version' row, simply treat as version 0 to insert the row
			return 0, nil
		}
		logger.L.Errorw("Error when querying schema version", "error", err)
		return 0, err
	}
	logger.L.Debugw("Got database schema version", "version", schemaVersion)
	return schemaVersion, nil
}

func InitDB(conn *pgxpool.Conn) error {
	// migration could be long, so we use background context instead of the timeout version
	return conn.BeginFunc(context.Background(), func(dbTx pgx.Tx) error {
		versionSqlMap, codeSchemaVersion, err := schema.GetVersionSQLMap()
		if err != nil {
			return err
		}
		dbSchemaVersion, err := GetSchemaVersion(dbTx)
		if err != nil {
			return err
		}
		for version := dbSchemaVersion + 1; version <= codeSchemaVersion; version++ {
			logger.L.Infow("Running SQL migration", "from_schema_version", version-1, "to_schema_version", version)
			sql := versionSqlMap[version]
			_, err := dbTx.Exec(context.Background(), sql)
			if err != nil {
				return err
			}
			_, err = dbTx.Exec(context.Background(), `UPDATE meta SET height = $1 WHERE id = 'schema_version'`, version)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
