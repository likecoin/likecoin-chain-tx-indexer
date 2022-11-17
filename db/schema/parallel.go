package schema

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
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

func checkBatchSize(batchSize uint64) error {
	if batchSize == 0 {
		return fmt.Errorf("invalid batch size (%d)", batchSize)
	}
	return nil
}

func MigrateSetupIscnVersionTable(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	err = checkMinSchemaVersion(conn, 7)
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

func MigrateIscnOwner(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating ISCN owner addresses")
	prevBatchLastPkey := int64(0)
	pkeyIds := make([]int64, batchSize)
	owners := make([]string, batchSize)
	for {
		rows, err := conn.Query(context.Background(), `
			SELECT id, owner
			FROM iscn
			WHERE id > $1
			ORDER BY id
			LIMIT $2
		`, prevBatchLastPkey, batchSize)
		if err != nil {
			logger.L.Debugw("Error when querying batch", "error", err)
			return err
		}
		count := 0
		for rows.Next() {
			err = rows.Scan(&pkeyIds[count], &owners[count])
			if err != nil {
				logger.L.Debugw("Error when scanning row", "error", err)
				return err
			}
			count++
		}
		if count == 0 {
			break
		}
		batch := pgx.Batch{}
		for i := 0; i < count; i++ {
			pkeyId := pkeyIds[i]
			convertedOwner, err := utils.ConvertAddressPrefix(owners[i], db.MainAddressPrefix)
			if err != nil {
				logger.L.Warnw(
					"Cannot convert ISCN owner address",
					"pkey_id", pkeyId,
					"owner", owners[i],
					"error", err,
				)
				continue
			}
			batch.Queue(
				`UPDATE iscn SET owner = $1 WHERE id = $2`,
				convertedOwner, pkeyId,
			)
		}
		if batch.Len() > 0 {
			results := conn.SendBatch(context.Background(), &batch)
			_, err = results.Exec()
			if err != nil {
				logger.L.Debugw(
					"Error when updating ISCN owner",
					"batch_first_pkey_id", pkeyIds[0],
					"batch_last_pkey_id", pkeyIds[count-1],
					"error", err,
				)
				return err
			}
			results.Close()
		}
		prevBatchLastPkey = pkeyIds[count-1]
		logger.L.Infow(
			"ISCN owner migration progress",
			"pkey_id", prevBatchLastPkey,
		)
	}
	logger.L.Info("Migration for ISCN owner addresses done")
	return nil
}

func MigrateNftOwner(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating NFT owner addresses")
	prevBatchLastPkey := int64(0)
	pkeyIds := make([]int64, batchSize)
	owners := make([]string, batchSize)
	for {
		rows, err := conn.Query(context.Background(), `
			SELECT id, owner
			FROM nft
			WHERE id > $1
			ORDER BY id
			LIMIT $2
		`, prevBatchLastPkey, batchSize)
		if err != nil {
			logger.L.Debugw("Error when querying batch", "error", err)
			return err
		}
		count := 0
		for rows.Next() {
			err = rows.Scan(&pkeyIds[count], &owners[count])
			if err != nil {
				logger.L.Debugw("Error when scanning row", "error", err)
				return err
			}
			count++
		}
		if count == 0 {
			break
		}
		batch := pgx.Batch{}
		for i := 0; i < count; i++ {
			pkeyId := pkeyIds[i]
			convertedOwner, err := utils.ConvertAddressPrefix(owners[i], db.MainAddressPrefix)
			if err != nil {
				logger.L.Warnw(
					"Cannot convert NFT owner address",
					"pkey_id", pkeyId,
					"owner", owners[i],
					"error", err,
				)
				continue
			}
			batch.Queue(
				`UPDATE nft SET owner = $1 WHERE id = $2`,
				convertedOwner, pkeyId,
			)
		}
		if batch.Len() > 0 {
			results := conn.SendBatch(context.Background(), &batch)
			_, err = results.Exec()
			if err != nil {
				logger.L.Debugw(
					"Error when updating NFT owner",
					"batch_first_pkey_id", pkeyIds[0],
					"batch_last_pkey_id", pkeyIds[count-1],
					"error", err,
				)
				return err
			}
			results.Close()
		}
		prevBatchLastPkey = pkeyIds[count-1]
		logger.L.Infow(
			"NFT owner migration progress",
			"pkey_id", prevBatchLastPkey,
		)
	}
	logger.L.Info("Migration for NFT owner addresses done")
	return nil
}

func MigrateNftEventSenderAndReceiver(conn *pgxpool.Conn, batchSize uint64) error {
	err := checkBatchSize(batchSize)
	if err != nil {
		return err
	}
	logger.L.Info("Start migrating NFT event sender and receiver addresses")
	prevBatchLastPkey := int64(0)
	pkeyIds := make([]int64, batchSize)
	senders := make([]string, batchSize)
	receivers := make([]string, batchSize)
	for {
		rows, err := conn.Query(context.Background(), `
			SELECT id, sender, receiver
			FROM nft_event
			WHERE id > $1
			ORDER BY id
			LIMIT $2
		`, prevBatchLastPkey, batchSize)
		if err != nil {
			logger.L.Debugw("Error when querying batch", "error", err)
			return err
		}
		count := 0
		for rows.Next() {
			err = rows.Scan(&pkeyIds[count], &senders[count], &receivers[count])
			if err != nil {
				logger.L.Debugw("Error when scanning row", "error", err)
				return err
			}
			count++
		}
		if count == 0 {
			break
		}
		batch := pgx.Batch{}
		for i := 0; i < count; i++ {
			pkeyId := pkeyIds[i]
			convertedSender, err := utils.ConvertAddressPrefix(senders[i], db.MainAddressPrefix)
			if err != nil {
				logger.L.Warnw(
					"Cannot convert NFT event sender address",
					"pkey_id", pkeyId,
					"sender", senders[i],
					"error", err,
				)
				convertedSender = senders[i]
			}
			convertedReceiver, err := utils.ConvertAddressPrefix(receivers[i], db.MainAddressPrefix)
			if err != nil {
				logger.L.Warnw(
					"Cannot convert NFT event receiver address",
					"pkey_id", pkeyId,
					"receiver", receivers[i],
					"error", err,
				)
				convertedReceiver = receivers[i]
			}
			batch.Queue(
				`UPDATE nft_event
				 SET
					sender = $1,
					receiver = $2
				WHERE id = $3`,
				convertedSender, convertedReceiver, pkeyId,
			)
		}
		if batch.Len() > 0 {
			results := conn.SendBatch(context.Background(), &batch)
			_, err = results.Exec()
			if err != nil {
				logger.L.Debugw(
					"Error when updating NFT event sender and receiver",
					"batch_first_pkey_id", pkeyIds[0],
					"batch_last_pkey_id", pkeyIds[count-1],
					"error", err,
				)
				return err
			}
			results.Close()
		}
		prevBatchLastPkey = pkeyIds[count-1]
		logger.L.Infow(
			"NFT event sender and receiver migration progress",
			"pkey_id", prevBatchLastPkey,
		)
	}
	logger.L.Info("Migration for NFT event sender and receiver addresses done")
	return nil
}

func MigrateAddressPrefix(conn *pgxpool.Conn, batchSize uint64) (err error) {
	err = MigrateIscnOwner(conn, batchSize)
	if err != nil {
		return err
	}
	err = MigrateNftOwner(conn, batchSize)
	if err != nil {
		return err
	}
	err = MigrateNftEventSenderAndReceiver(conn, batchSize)
	if err != nil {
		return err
	}
	return nil
}
