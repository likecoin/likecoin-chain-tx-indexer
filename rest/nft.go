package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleNftClass(c *gin.Context) {
	var q db.QueryClassRequest

	if err := c.ShouldBindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err})
		return
	}

	conn := getConn(c)
	res, err := db.GetClasses(conn, q, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNft(c *gin.Context) {
	var q db.QueryNftRequest

	if err := c.ShouldBindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn := getConn(c)
	res, err := db.GetNfts(conn, q, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftOwner(c *gin.Context) {
	var q db.QueryOwnerRequest

	if err := c.ShouldBindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid inputs: " + err.Error()})
		return
	}

	conn := getConn(c)
	res, err := db.GetOwners(conn, q)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftEvents(c *gin.Context) {
	var form db.QueryEventsRequest
	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid inputs: " + err.Error()})
		return
	}

	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	if form.ClassId == "" && form.IscnIdPrefix == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "must provide class_id or iscn_id_prefix"})
		return
	}
	conn := getConn(c)

	res, err := db.GetNftEvents(conn, form, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftRanking(c *gin.Context) {
	var q db.QueryRankingRequest
	if err := c.ShouldBind(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn := getConn(c)
	res, err := db.GetClassesRanking(conn, q)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftCollectors(c *gin.Context) {
	var form db.QueryCollectorRequest
	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid inputs: " + err.Error()})
		return
	}

	conn := getConn(c)
	res, err := db.GetCollector(conn, form)

	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftCreators(c *gin.Context) {
	var form db.QueryCreatorRequest
	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid inputs: " + err.Error()})
		return
	}

	conn := getConn(c)
	res, err := db.GetCreators(conn, form)

	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
