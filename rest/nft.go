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

	c.AbortWithStatusJSON(400, gin.H{"error": "params is require"})
}

func handleNftByIscn(c *gin.Context) {
	q := c.Request.URL.Query()

	iscn := q.Get("iscn_id_prefix")
	conn := getConn(c)
	res, err := db.GetNftByIscn(conn, iscn)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	c.JSON(200, res)
}
