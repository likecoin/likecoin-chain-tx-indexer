package extractor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

var ExtractFunc db.Extractor

// TODO: should we make extractor synchronous with poller instead of async?
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
			finished, err = db.Extract(conn, ExtractFunc)
			if err != nil {
				logger.L.Errorw("Extract error", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}
			if finished {
				height := <-trigger
				logger.L.Debugf("Extractor: trigger by poller on height %d", height)
			}
		}
	}()
	return trigger
}

func init() {
	eventExtractor := NewEventExtractor()

	// TODO: rewrite all extractors so they don't rely on message.action (which won't be present in authz exec)
	eventExtractor.Register("message", "action", "create_iscn_record", insertIscn)
	eventExtractor.Register("message", "action", "/likechain.iscn.MsgCreateIscnRecord", insertIscn)
	eventExtractor.Register("message", "action", "update_iscn_record", insertIscn)
	eventExtractor.Register("message", "action", "/likechain.iscn.MsgUpdateIscnRecord", insertIscn)
	eventExtractor.Register("message", "action", "msg_change_iscn_record_ownership", transferIscn)
	eventExtractor.Register("message", "action", "/likechain.iscn.MsgChangeIscnRecordOwnership", transferIscn)
	eventExtractor.Register("message", "action", "new_class", createNftClass)
	eventExtractor.Register("message", "action", "update_class", updateNftClass)
	eventExtractor.Register("message", "action", "mint_nft", mintNft)
	eventExtractor.Register("message", "action", "/cosmos.nft.v1beta1.MsgSend", sendNft)
	eventExtractor.Register("message", "action", "buy_nft", buyNft)
	eventExtractor.Register("message", "action", "sell_nft", sellNft)
	eventExtractor.Register("message", "action", "create_listing", createListing)
	eventExtractor.Register("message", "action", "update_listing", updateListing)
	eventExtractor.Register("message", "action", "delete_listing", deleteListing)
	eventExtractor.Register("message", "action", "create_offer", createOffer)
	eventExtractor.Register("message", "action", "update_offer", updateOffer)
	eventExtractor.Register("message", "action", "delete_offer", deleteOffer)

	ExtractFunc = eventExtractor.Extract
}
