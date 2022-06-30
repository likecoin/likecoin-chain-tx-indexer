package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	iscnTypes "github.com/likecoin/likecoin-chain/v2/x/iscn/types"
)

type ISCNRecordsResponse struct {
	Records []iscnTypes.QueryResponseRecord `json:"records"`
}

func handleISCN(c *gin.Context) {
	q := c.Request.URL.Query()
	if q.Get("q") != "" {
		handleISCNSearch(c)
		return
	}

	var iscn db.ISCN
	hasQuery := false

	for k, v := range q {
		log.Println(k, v)
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
		case "limit", "page":
		default:
			c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("unknown query %s", k)})
			return
		}
	}

	p := getPagination(q)

	pool := getPool(c)
	var records []iscnTypes.QueryResponseRecord
	var err error
	if hasQuery {
		records, err = db.QueryISCN(pool, iscn, p)
	} else {
		records, err = db.QueryISCNList(pool, p)
	}
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	respondRecords(c, records)
}

func handleISCNSearch(c *gin.Context) {
	q := c.Request.URL.Query()
	p := getPagination(q)
	term := q.Get("q")
	if term == "" {
		c.AbortWithStatusJSON(404, gin.H{"error": "parameter 'q' is required"})
		return
	}
	pool := getPool(c)
	records, err := db.QueryISCNAll(pool, term, p)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	respondRecords(c, records)
}

func getPagination(q url.Values) db.Pagination {
	p := db.Pagination{
		Limit: 1,
		Page:  0,
		Order: db.ORDER_DESC,
	}
	if page, err := getPage(q, "page"); err == nil {
		p.Page = page
	}
	if limit, err := getLimit(q, "limit"); err == nil {
		p.Limit = limit
	}
	return p
}

func respondRecords(c *gin.Context, iscnInputs []iscnTypes.QueryResponseRecord) {
	if len(iscnInputs) == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "Record not found"})
		return
	}

	response := ISCNRecordsResponse{
		Records: iscnInputs,
	}
	resJson, err := json.Marshal(&response)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(resJson)
}
