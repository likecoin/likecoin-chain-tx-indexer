package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleNftMarketplaceItem(c *gin.Context) {
	var q db.QueryNftMarketplaceItemsRequest

	if err := c.ShouldBindQuery(&q); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	if q.Type != "listing" && q.Type != "offer" {
		c.AbortWithStatusJSON(400, gin.H{"error": `invalid type (expect "listing" or "offer")`})
		return
	}

	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err})
		return
	}

	conn := getConn(c)
	res, err := db.GetNftMarketplaceItems(conn, q, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
