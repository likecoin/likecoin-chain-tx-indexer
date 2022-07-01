package db

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func GetISCNTxs(conn *pgxpool.Conn, begin, end int64) (pgx.Rows, error) {
	ctx, _ := GetTimeoutContext()

	sql := `
		SELECT height, tx #> '{"tx", "body", "messages", 0, "record"}' as record, events, tx #> '{"timestamp"}'
		FROM txs
		WHERE height >= $1
			AND height < $2
			AND events @> '{"message.module=\"iscn\""}'
		ORDER BY height ASC;
	`
	return conn.Query(ctx, sql, begin, end)
}

func GetISCNHeight(conn *pgxpool.Conn) (int64, error) {
	ctx, _ := GetTimeoutContext()
	var height int64
	err := conn.QueryRow(ctx, `SELECT height FROM meta WHERE id = 'iscn'`).Scan(&height)
	return height, err
}

func (batch *Batch) InsertISCN(iscn ISCN) {
	sql := `
INSERT INTO iscn (iscn_id, iscn_id_prefix, owner, keywords, fingerprints, stakeholders, data, timestamp, ipld) VALUES
( $1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT DO NOTHING;`
	batch.Batch.Queue(sql, iscn.Iscn, iscn.IscnPrefix, iscn.Owner,
		iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders, iscn.Data, iscn.Timestamp, iscn.Ipld)
}

func (batch *Batch) TransferISCN(events types.StringEvents) {
	sender := utils.GetEventsValue(events, "message", "sender")
	iscnId := utils.GetEventsValue(events, "iscn_record", "iscn_id")
	newOwner := utils.GetEventsValue(events, "iscn_record", "owner")
	batch.Batch.Queue(`UPDATE iscn SET owner = $2 WHERE iscn_id = $1`, iscnId, newOwner)
	logger.L.Infof("Send ISCN %s from %s to %s\n", iscnId, sender, newOwner)
}

func (batch *Batch) UpdateISCNHeight(height int64) {
	batch.Batch.Queue(`UPDATE meta SET height = $1 WHERE id = 'iscn'`, height)
}
