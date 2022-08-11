package db

import (
	"encoding/json"

	"github.com/jackc/pgx/v4/pgxpool"
)

func GetNftMintCount(conn *pgxpool.Conn, q QueryNftMintCountRequest) (count uint64, err error) {
	sql := `
	SELECT COUNT(n.id)
	FROM nft as n
	JOIN nft_class AS c USING (class_id)
	JOIN iscn as i ON c.parent_iscn_id_prefix = i.iscn_id_prefix
	WHERE ($1 = true OR i.owner != n.owner)
		AND ($2::text[] IS NULL OR n.owner != ALL($2))
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	err = conn.QueryRow(ctx, sql, q.IncludeOwner, q.IgnoreList).Scan(&count)
	if err != nil {
		panic(err)
	}
	return
}

func GetNftTradeStats(conn *pgxpool.Conn, q QueryNftTradeStatsRequest) (res QueryNftTradeStatsResponse, err error) {
	payload := struct {
		Type    string `json:"@type"`
		Grantee string `json:"grantee"`
	}{
		"/cosmos.authz.v1beta1.MsgExec",
		q.ApiAddress,
	}
	payloadJSON, err := json.Marshal(&payload)
	if err != nil {
		panic(err)
	}

	sql := `
	SELECT count(*), sum((tx #>> '{"tx", "body", "messages", 0, "msgs", 0, "amount", 0, "amount"}')::bigint)
	from txs
	WHERE tx #> '{"tx", "body", "messages", 0}' @> $1
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err = conn.QueryRow(ctx, sql, string(payloadJSON)).Scan(
		&res.Count, &res.TotalVolume,
	)
	if err != nil {
		panic(err)
	}
	return
}

func GetMintNftWalletCount() {

}

func GetOwnNftWalletCount() {

}

func GetWalletCollectionList() {

}

func GetTransactionValue() {

}
