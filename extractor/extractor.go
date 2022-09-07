package extractor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

var handlers = map[string]db.EventHandler{
	"create_iscn_record":                           insertIscn,
	"/likechain.iscn.MsgCreateIscnRecord":          insertIscn,
	"update_iscn_record":                           insertIscn,
	"/likechain.iscn.MsgUpdateIscnRecord":          insertIscn,
	"msg_change_iscn_record_ownership":             transferIscn,
	"/likechain.iscn.MsgChangeIscnRecordOwnership": transferIscn,
	"new_class":                   createNftClass,
	"update_class":                updateNftClass,
	"mint_nft":                    mintNft,
	"/cosmos.nft.v1beta1.MsgSend": sendNft,
}

func Run(pool *pgxpool.Pool) chan<- int64 {
	trigger := make(chan int64, 100)
	go func() {
		conn, err := db.AcquireFromPool(pool)
		if err != nil {
			logger.L.Errorw("Failed to acquire connection for extractor", "error", err)
			return
		}

		logger.L.Info("Extractor started")
		var finished bool
		for {
			if err = conn.Ping(context.Background()); err != nil {
				conn, err = db.AcquireFromPool(pool)
				if err != nil {
					logger.L.Errorw("Failed to acquire connection for extractor", "error", err)
					time.Sleep(10 * time.Second)
					continue
				}
			}
			finished, err = db.Extract(conn, handlers)
			if err != nil {
				logger.L.Errorw("Extract error", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}
			if finished {
				height := <-trigger
				logger.L.Infof("Extractor: trigger by poller on height %d", height)
			}
		}
	}()
	return trigger
}
