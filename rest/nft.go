package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleNftClass(c *gin.Context) {
	var q db.QueryClassRequest

	if err := c.BindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn := getConn(c)
	res, err := db.GetClasses(conn, q)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNft(c *gin.Context) {
	var q db.QueryNftRequest

	if err := c.BindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn := getConn(c)
	res, err := db.GetNftByOwner(conn, q)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftOwner(c *gin.Context) {
	q := c.Request.URL.Query()

	classId := q.Get("class_id")
	if classId == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "class_id is require"})
		return
	}

	conn := getConn(c)
	res, err := db.GetOwnerByClassId(conn, classId)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftEvents(c *gin.Context) {
	var form db.QueryEventsRequest
	c.BindQuery(&form)

	if form.ClassId == "" && form.IscnIdPrefix == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "must provide class_id or iscn_id_prefix"})
		return
	}
	conn := getConn(c)

	res, err := db.GetNftEvents(conn, form)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
