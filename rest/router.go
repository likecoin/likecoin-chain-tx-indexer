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
const ANALYSIS_ENDPOINT = "/statistics"
const INFO_ENDPOINT = "/indexer/info"

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
	_ = router.Run(listenAddr)
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
		nft.GET("/income", handleNftIncome)
		nft.GET("/user-stat", handleNftUserStat)
		nft.GET("/marketplace", handleNftMarketplaceItem)
		nft.GET("/collector-top-ranked-creators", handleNftCollectorTopRankedCreatorsRequest)
		nft.GET("/classes-owners", handleClassesOwnersRequest)
	}
	analysis := router.Group(ANALYSIS_ENDPOINT)
	{
		analysis.GET("/iscn/record-count", handleISCNRecordCount)
		analysis.GET("/iscn/owner-count", handleISCNOwnerCount)
		analysis.GET("/nft/nft-count", handleNftCount)
		analysis.GET("/nft/trade", handleNftTradeStats)
		analysis.GET("/nft/creator-count", handleNftCreatorCount)
		analysis.GET("/nft/returning-creator-count", handleNftRecentCreatorCount)
		analysis.GET("/nft/owner-count", handleNftOwnerCount)
		analysis.GET("/nft/owners", handleNftOwnerList)
	}
	router.Use(filterContentFingerprints())
	router.GET(ISCN_ENDPOINT, handleIscn)
	router.GET(STARGATE_ENDPOINT, handleStargateTxsSearch)
	router.GET(LATEST_HEIGHT_ENDPOINT, handleLatestHeight)
	router.GET(INFO_ENDPOINT, handleInfo)
	return router
}

func with(key string, value interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(key, value)
		c.Next()
	}
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
