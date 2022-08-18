package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleIscn(c *gin.Context) {
	q := c.Request.URL.Query()
	if q.Get("q") != "" {
		handleIscnSearch(c)
		return
	}

	var form db.IscnQuery

	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn := getConn(c)
	var res db.IscnResponse
	if form.Empty() {
		res, err = db.QueryIscnList(conn, p)
	} else {
		res, err = db.QueryIscn(conn, form, p)
	}
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleIscnSearch(c *gin.Context) {
	q := c.Request.URL.Query()
	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	term := q.Get("q")
	if term == "" {
		c.AbortWithStatusJSON(404, gin.H{"error": "parameter 'q' is required"})
		return
	}
	conn := getConn(c)
	res, err := db.QueryIscnSearch(conn, term, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
