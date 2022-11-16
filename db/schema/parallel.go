package schema

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func checkMinSchemaVersion(conn *pgxpool.Conn, minVersion uint64) error {
	version, err := GetSchemaVersion(conn)
	if err != nil {
		return err
	}
	if version < minVersion {
		logger.L.Errorw("Schema version does not meet minimum requirement", "minimum_version", minVersion, "current_version", version)
		return fmt.Errorf("schema version too low")
	}
	return nil
}

func MigrationSetupIscnVersionTable(conn *pgxpool.Conn, batchSize uint64) error {
	if batchSize == 0 {
		return fmt.Errorf("invalid batch size (%d)", batchSize)
	}
	err := checkMinSchemaVersion(conn, 7)
	if err != nil {
		return err
	}
	return conn.BeginFunc(context.Background(), func(dbTx pgx.Tx) error {
		_, err := dbTx.Exec(context.Background(), `
			DECLARE iscn_version_migration_cursor CURSOR FOR
				SELECT id, iscn_id_prefix, version
				FROM iscn
				ORDER BY id
			;
		`)
		if err != nil {
			logger.L.Debugw("Error when declaring cursor", "error", err)
			return err
		}
		var pkeyId int64
		iscnIdPrefixes := make([]string, batchSize)
		versions := make([]int64, batchSize)
		for {
			rows, err := dbTx.Query(context.Background(), fmt.Sprintf(`FETCH %d FROM iscn_version_migration_cursor`, batchSize))
			if err != nil {
				logger.L.Debugw("Error when fetching from cursor", "error", err)
				return err
			}
			defer rows.Close()
			count := 0
			for rows.Next() {
				err = rows.Scan(&pkeyId, &iscnIdPrefixes[count], &versions[count])
				if err != nil {
					logger.L.Debugw("Error when scanning row", "error", err)
					return err
				}
				count++
			}
			if count == 0 {
				break
			}
			for i := 0; i < count; i++ {
				_, err = dbTx.Exec(context.Background(), `
						INSERT INTO iscn_latest_version AS t (iscn_id_prefix, latest_version)
						VALUES ($1, $2)
						ON CONFLICT (iscn_id_prefix) DO UPDATE
							SET latest_version = GREATEST(t.latest_version, EXCLUDED.latest_version)
						;
					`, iscnIdPrefixes[i], versions[i])
				if err != nil {
					logger.L.Debugw("Error when inserting into iscn_latest_version", "error", err)
					return err
				}
			}
			logger.L.Infow(
				"ISCN version table migration progress",
				"pkey_id", pkeyId,
				"iscn_id_prefix", iscnIdPrefixes[count-1],
				"version", versions[count-1],
			)
		}
		return nil
	})
}

func MigrateOwnerAddressDualPrefix(conn *pgxpool.Conn) error {
	return nil // TODO
}
