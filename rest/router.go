package rest

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

const STARGATE_ENDPOINT = "/cosmos/tx/v1beta1/txs"
const ISCN_ENDPOINT = "/iscn/records"
const LATEST_HEIGHT_ENDPOINT = "/indexer/latest/height"

func Run(pool *pgxpool.Pool, listenAddr string, lcdEndpoint string) {
	lcdURL, err := url.Parse(lcdEndpoint)
	if err != nil {
		logger.L.Panicw("Cannot parse lcd URL", "lcd_endpoint", lcdEndpoint, "error", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(lcdURL)
	proxyHandler := func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	router := getRouter(pool)
	router.NoRoute(proxyHandler)
	router.Run(listenAddr)
}

func getRouter(pool *pgxpool.Pool) *gin.Engine {
	router := gin.New()
	router.GET(ISCN_ENDPOINT, withPool(pool), handleISCN)
	router.GET("/txs", withConn(pool), handleAminoTxsSearch)
	router.GET(STARGATE_ENDPOINT, withConn(pool), handleStargateTxsSearch)
	router.GET(LATEST_HEIGHT_ENDPOINT, handleLatestHeight)
	return router
}

func withConn(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := db.AcquireFromPool(pool)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Set("conn", conn)
		defer conn.Release()
		c.Next()
	}
}

func withPool(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("pool", pool)
		c.Next()
	}
}
