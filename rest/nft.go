package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleNft(c *gin.Context) {
	q := c.Request.URL.Query()
	if q.Get("iscn_id_prefix") != "" {
		handleNftByIscn(c)
		return
	}
	if q.Get("owner") != "" {
		handleNftByOwner(c)
		return
	}
	if q.Get("class_id") != "" {
		handleOwnerByClassId(c)
		return
	}

	c.AbortWithStatusJSON(400, gin.H{"error": "params is require"})
}

func handleNftByIscn(c *gin.Context) {
	q := c.Request.URL.Query()

	iscn := q.Get("iscn_id_prefix")
	conn := getConn(c)
	res, err := db.GetNftByIscn(conn, iscn)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}

	c.JSON(200, res)
}

func handleNftByOwner(c *gin.Context) {
	q := c.Request.URL.Query()

	owner := q.Get("owner")
	if owner == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "owner is require"})
		return
	}

	conn := getConn(c)
	res, err := db.GetNftByOwner(conn, owner)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}

	c.JSON(200, res)
}

func handleOwnerByClassId(c *gin.Context) {
	q := c.Request.URL.Query()

	classId := q.Get("class_id")

	conn := getConn(c)
	res, err := db.GetOwnerByClassId(conn, classId)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}

	c.JSON(200, res)
}

func handleNftEvents(c *gin.Context) {
	q := c.Request.URL.Query()

	classId := q.Get("class_id")
	if classId == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "class_id is required"})
		return
	}
	nftId := q.Get("nft_id")

	conn := getConn(c)
	res, err := db.GetNftEventsByNftId(conn, classId, nftId)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}

	c.JSON(200, res)
}
