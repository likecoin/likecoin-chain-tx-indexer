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
const LATEST_HEIGHT_ENDPOINT = "/indexer/height/latest"
const NFT_ENDPOINT = "/likechain/likenft/v1"
const ANALYSIS_ENDPOINT = "/statistics/nft"

func Run(pool *pgxpool.Pool, listenAddr string, lcdEndpoint string, defaultApiAddresses []string) {
	lcdURL, err := url.Parse(lcdEndpoint)
	if err != nil {
		logger.L.Panicw("Cannot parse lcd URL", "lcd_endpoint", lcdEndpoint, "error", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(lcdURL)
	proxyHandler := func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	router := GetRouter(pool, defaultApiAddresses)
	router.NoRoute(proxyHandler)
	router.Run(listenAddr)
}

func GetRouter(pool *pgxpool.Pool, defaultApiAddresses []string) *gin.Engine {
	router := gin.New()
	router.Use(withConn(pool), withDefaultApiAddresses(defaultApiAddresses))
	nft := router.Group(NFT_ENDPOINT)
	{
		nft.GET("/class", handleNftClass)
		nft.GET("/nft", handleNft)
		nft.GET("/owner", handleNftOwner)
		nft.GET("/event", handleNftEvents)
		nft.GET("/ranking", handleNftRanking)
		nft.GET("/collector", handleNftCollectors)
		nft.GET("/creator", handleNftCreators)
		nft.GET("/user-stat", handleNftUserStat)
	}
	analysis := router.Group(ANALYSIS_ENDPOINT)
	{
		analysis.GET("/nft-count", handleNftCount)
		analysis.GET("/trade", handleNftTradeStats)
		analysis.GET("/creator-count", handleNftCreatorCount)
		analysis.GET("/owner-count", handleNftOwnerCount)
		analysis.GET("/owners", handleNftOwnerList)
	}
	router.GET(ISCN_ENDPOINT, handleIscn)
	router.GET("/txs", handleAminoTxsSearch)
	router.GET(STARGATE_ENDPOINT, handleStargateTxsSearch)
	router.GET(LATEST_HEIGHT_ENDPOINT, handleLatestHeight)
	return router
}

func with(key string, value interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(key, value)
		c.Next()
	}
}

func withPool(pool *pgxpool.Pool) gin.HandlerFunc {
	return with("pool", pool)
}

func getPool(c *gin.Context) *pgxpool.Pool {
	return c.MustGet("pool").(*pgxpool.Pool)
}

func withDefaultApiAddresses(defaultApiAddresses []string) gin.HandlerFunc {
	return with("default-api-addresses", defaultApiAddresses)
}

func getDefaultApiAddresses(c *gin.Context) []string {
	return c.MustGet("default-api-addresses").([]string)
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

func getConn(c *gin.Context) *pgxpool.Conn {
	return c.MustGet("conn").(*pgxpool.Conn)
}
