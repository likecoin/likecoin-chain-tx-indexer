package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleIscn(c *gin.Context) {
	var form db.IscnQuery

	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	if form.SearchTerm != "" {
		handleIscnSearch(c, form)
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
		res, err = db.QueryIscnList(conn, p, form.AllIscnVersions)
	} else {
		res, err = db.QueryIscn(conn, form, p)
	}
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleIscnSearch(c *gin.Context, form db.IscnQuery) {
	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	term := form.SearchTerm
	conn := getConn(c)
	res, err := db.QueryIscnSearch(conn, term, p, form.AllIscnVersions)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
