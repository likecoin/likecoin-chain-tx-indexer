package rest

import (
	"encoding/json"
	"fmt"

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

	var iscn db.ISCN
	hasQuery := false

	for k, v := range q {
		switch k {
		case "iscn_id":
			if len(v) > 0 {
				hasQuery = true
				iscn.Iscn = v[0]
			}
		case "owner":
			if len(v) > 0 {
				hasQuery = true
				iscn.Owner = v[0]
			}
		case "fingerprint", "fingerprints":
			hasQuery = true
			iscn.Fingerprints = v
		case "keywords":
			hasQuery = true
			iscn.Keywords = v
		case "stakeholders.entity.id", "stakeholders.id":
			if len(v) > 0 {
				hasQuery = true
				iscn.Stakeholders = []byte(fmt.Sprintf(`[{"id": "%s"}]`, v[0]))
			}
		case "stakeholders.entity.name", "stakeholders.name":
			if len(v) > 0 {
				hasQuery = true
				iscn.Stakeholders = []byte(fmt.Sprintf(`[{"name": "%s"}]`, v[0]))
			}
		}
	}

	p, err := getPagination(c)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	pool := getPool(c)
	var res db.ISCNResponse
	if hasQuery {
		res, err = db.QueryISCN(pool, iscn, p)
	} else {
		res, err = db.QueryISCNList(pool, p)
	}
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	respondRecords(c, res)
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
	pool := getPool(c)
	res, err := db.QueryISCNAll(pool, term, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	respondRecords(c, res)
}

func getPagination(c *gin.Context) (p db.PageRequest, err error) {
	p = db.PageRequest{}
	err = c.ShouldBindQuery(&p)
	logger.L.Debugf("%#v", p)
	return p, err
}

func respondRecords(c *gin.Context, res db.ISCNResponse) {
	if len(res.Records) == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "Record not found"})
		return
	}

	resJson, err := json.Marshal(&res)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(resJson)
}
