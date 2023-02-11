package parallel

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

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
