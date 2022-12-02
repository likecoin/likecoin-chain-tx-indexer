package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleNftCount(c *gin.Context) {
	var q db.QueryNftCountRequest
	if err := c.ShouldBindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	res, err := db.GetNftCount(getConn(c), q)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftTradeStats(c *gin.Context) {
	var q db.QueryNftTradeStatsRequest
	if err := c.ShouldBindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	if len(q.ApiAddresses) == 0 {
		q.ApiAddresses = getDefaultApiAddresses(c)
	}

	res, err := db.GetNftTradeStats(getConn(c), q)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftCreatorCount(c *gin.Context) {
	res, err := db.GetNftCreatorCount(getConn(c))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftOwnerCount(c *gin.Context) {
	res, err := db.GetNftOwnerCount(getConn(c))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftOwnerList(c *gin.Context) {
	var q db.PageRequest
	if err := c.ShouldBindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	res, err := db.GetNftOwnerList(getConn(c), q)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
