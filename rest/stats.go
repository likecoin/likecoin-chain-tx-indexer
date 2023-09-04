package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleISCNRecordCount(c *gin.Context) {
	res, err := db.GetISCNRecordCount(getConn(c))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleISCNOwnerCount(c *gin.Context) {
	res, err := db.GetISCNOwnerCount(getConn(c))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

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

func handleNftRecentCreatorCount(c *gin.Context) {
	var q db.QueryNftReturningCreatorCountRequest
	if err := c.ShouldBindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid inputs: " + err.Error()})
		return
	}

	if q.ReturningThresholdDays < 7 || q.ReturningThresholdDays > 365 {
		c.AbortWithStatusJSON(400, gin.H{"error": "returning_threshold_days should be between 7 and 365"})
		return
	}

	if q.Interval != "" && q.Interval != "week" && q.Interval != "month" {
		c.AbortWithStatusJSON(400, gin.H{"error": "interval should be 'week' or 'month'"})
		return
	}

	if q.After != 0 && q.Before != 0 && q.Before-q.After > 60*60*24*30*365.25 {
		c.AbortWithStatusJSON(400, gin.H{"error": "before - after should be less than 1 year"})
		return
	}

	res, err := db.GetNftReturningCreatorCount(getConn(c), q)
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
