package db

import (
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func GetNFTTxs(conn *pgxpool.Conn, begin, end int64) (pgx.Rows, error) {
	ctx, _ := GetTimeoutContext()
	sql := `
SELECT tx FROM txs
SELECT tx #> '{"tx", "body", "messages"}' AS messages, tx -> 'logs' AS logs FROM txs
where height >= $1
	AND height < $2
	AND events @> '{"message.action=\"new_class\""}'
ORDER BY height ASC
	`
	return conn.Query(ctx, sql, begin, end)
}
