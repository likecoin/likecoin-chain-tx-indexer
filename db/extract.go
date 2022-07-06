package db

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

func GetMetaHeight(conn *pgxpool.Conn, key string) (int64, error) {
	ctx, _ := GetTimeoutContext()
	var height int64
	err := conn.QueryRow(ctx, `SELECT height FROM meta WHERE id = $1`, key).Scan(&height)
	return height, err
}

func (batch *Batch) InsertISCN(iscn ISCN) {
	sql := `
INSERT INTO iscn (iscn_id, iscn_id_prefix, version, owner, keywords, fingerprints, stakeholders, data, timestamp, ipld) VALUES
( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT DO NOTHING;`
	batch.Batch.Queue(sql, iscn.Iscn, iscn.IscnPrefix, iscn.Version, iscn.Owner,
		iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders, iscn.Data, iscn.Timestamp, iscn.Ipld)
}

func (batch *Batch) UpdateISCNHeight(height int64) {
	batch.Batch.Queue(`UPDATE meta SET height = $1 WHERE id = 'iscn'`, height)
}
