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
	ExtractFunc = eventExtractor.Extract
}
