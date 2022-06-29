package poller

import (
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

var LIMIT = 10000

func PollISCN(pool *pgxpool.Pool) {
	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Release()

	var finished bool
	for !finished && err == nil {
		finished, err = db.ConvertISCN(conn, LIMIT)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Finished")
}
