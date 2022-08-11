package db

import "github.com/jackc/pgx/v4/pgxpool"

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

	debugSQL(conn, ctx, sql, q.IncludeOwner, q.IgnoreList)
	err = conn.QueryRow(ctx, sql, q.IncludeOwner, q.IgnoreList).Scan(&count)
	if err != nil {
		panic(err)
	}
	return
}

func GetNftTradeCount() {

}

func GetMintNftWalletCount() {

}

func GetOwnNftWalletCount() {

}

func GetWalletCollectionList() {

}

func GetTransactionValue() {

}
