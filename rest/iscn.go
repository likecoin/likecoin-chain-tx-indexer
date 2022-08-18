package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func handleISCN(c *gin.Context) {
	q := c.Request.URL.Query()
	if q.Get("q") != "" {
		handleISCNSearch(c)
		return
	}

	var form db.ISCNQuery

	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.L.Debugw("", "form", form)

	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn := getConn(c)
	var res db.ISCNResponse
	if form.Empty() {
		res, err = db.QueryISCNList(conn, p)
	} else {
		res, err = db.QueryISCN(conn, form, p)
	}
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleISCNSearch(c *gin.Context) {
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
	res, err := db.QueryISCNAll(conn, term, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
