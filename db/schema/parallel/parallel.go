package parallel

import (
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/likecoin/likecoin-chain-tx-indexer/db/schema"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func checkMinSchemaVersion(conn *pgxpool.Conn, minVersion uint64) error {
	version, err := schema.GetSchemaVersion(conn)
	if err != nil {
		return err
	}
	if version < minVersion {
		logger.L.Errorw("Schema version does not meet minimum requirement", "minimum_version", minVersion, "current_version", version)
		return fmt.Errorf("schema version too low, expect at least %d, got %d", minVersion, version)
	}
	return nil
}

func checkBatchSize(batchSize uint64) error {
	if batchSize == 0 {
		return fmt.Errorf("invalid batch size (%d)", batchSize)
	}
	return nil
}
