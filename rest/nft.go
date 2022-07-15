package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleNftByIscn(c *gin.Context) {
	q := c.Request.URL.Query()

	iscn := q.Get("iscn_id_prefix")
	if iscn == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "iscn_id_prefix is required"})
		return
	}
	expand, err := getBool(q, "expand")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err})
		return
	}

	conn := getConn(c)
	res, err := db.GetNftByIscn(conn, iscn, expand)
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
	if classId == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "class_id is require"})
		return
	}

	conn := getConn(c)
	res, err := db.GetOwnerByClassId(conn, classId)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
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
