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

	if form.ClassId == "" &&
		form.IscnIdPrefix == "" &&
		len(form.Sender) == 0 &&
		len(form.Receiver) == 0 &&
		len(form.Creator) == 0 &&
		len(form.Involver) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "must provide either class_id, iscn_id_prefix, sender, receiver, creator or involver"})
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

	if q.OrderBy != "" && q.OrderBy != "total_sold_value" && q.OrderBy != "sold_count" {
		c.AbortWithStatusJSON(400, gin.H{"error": "order_by should either be total_sold_value or sold_count"})
		return
	}

	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err})
		return
	}

	if len(q.ApiAddresses) == 0 {
		q.ApiAddresses = getDefaultApiAddresses(c)
	}

	conn := getConn(c)
	res, err := db.GetClassesRanking(conn, q, p)
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
	if form.PriceBy != "nft" && form.PriceBy != "class" {
		c.AbortWithStatusJSON(400, gin.H{"error": "price_by should either be nft or class"})
		return
	}
	if form.OrderBy != "price" && form.OrderBy != "count" {
		c.AbortWithStatusJSON(400, gin.H{"error": "order_by should either be price or count"})
		return
	}
	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err})
		return
	}

	conn := getConn(c)

	res, err := db.GetCollector(conn, form, p)
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
	if form.PriceBy != "nft" && form.PriceBy != "class" {
		c.AbortWithStatusJSON(400, gin.H{"error": "price_by should either be nft or class"})
		return
	}
	if form.OrderBy != "price" && form.OrderBy != "count" {
		c.AbortWithStatusJSON(400, gin.H{"error": "order_by should either be price or count"})
		return
	}
	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err})
		return
	}

	conn := getConn(c)

	res, err := db.GetCreators(conn, form, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftUserStat(c *gin.Context) {
	var form db.QueryUserStatRequest
	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid inputs: " + err.Error()})
		return
	}

	conn := getConn(c)

	res, err := db.GetUserStat(conn, form)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleNftCollectorTopRankedCreatorsRequest(c *gin.Context) {
	var form db.QueryCollectorTopRankedCreatorsRequest
	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid inputs: " + err.Error()})
		return
	}

	conn := getConn(c)
	res, err := db.GetCollectorTopRankedCreators(conn, form)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func handleClassesOwnedRequest(c *gin.Context) {
	var form db.QueryClassesOwnedRequest
	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid inputs: " + err.Error()})
		return
	}
	conn := getConn(c)
	res, err := db.GetClassesOwned(conn, form)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
