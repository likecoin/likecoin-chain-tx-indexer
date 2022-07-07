package extractor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

var handlers = map[string]db.EventHandler{
	"create_iscn_record":                           insertISCN,
	"/likechain.iscn.MsgCreateIscnRecord":          insertISCN,
	"update_iscn_record":                           insertISCN,
	"/likechain.iscn.MsgUpdateIscnRecord":          insertISCN,
	"msg_change_iscn_record_ownership":             transferISCN,
	"/likechain.iscn.MsgChangeIscnRecordOwnership": transferISCN,
	"new_class": createNftClass,
	"mint_nft":  mintNft,
}

func Run(pool *pgxpool.Pool) chan<- int64 {
	trigger := make(chan int64, 100)
	go func() {
		conn, err := db.AcquireFromPool(pool)
		if err != nil {
			logger.L.Errorw("Failed to acquire connection for ISCN extractor", "error", err)
			return
		}

		logger.L.Info("ISCN extractor started")
		var finished bool
		for {
			if err = conn.Ping(context.Background()); err != nil {
				conn, err = db.AcquireFromPool(pool)
				if err != nil {
					logger.L.Errorw("Failed to acquire connection for ISCN extractor", "error", err)
					time.Sleep(10 * time.Second)
					continue
				}
			}
			finished, err = db.Extract(conn, handlers)
			if err != nil {
				logger.L.Errorw("Extract ISCN error", "error", err)
				time.Sleep(10 * time.Second)
				continue
			}
			if finished {
				height := <-trigger
				logger.L.Infof("ISCN extractor: trigger by poller, sync to %d", height)
			}
		}
	}()
	return trigger
}
