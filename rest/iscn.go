package rest

import (
	"encoding/json"
	"net/url"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	iscnTypes "github.com/likecoin/likecoin-chain/v3/x/iscn/types"
)

type ISCNRecordsResponse struct {
	Records []iscnTypes.QueryResponseRecord `json:"records"`
}

func handleISCN(c *gin.Context) {
	q := c.Request.URL.Query()
	hasQuery := false

	if q.Get("q") != "" {
		handleISCNSearch(c)
		return
	}

	events := make([]types.StringEvent, 0)
	if owner := q.Get("owner"); owner != "" {
		events = append(events, types.StringEvent{
			Type: "iscn_record",
			Attributes: []types.Attribute{
				{
					Key:   "owner",
					Value: owner,
				},
			},
		})
		hasQuery = true
	}
	if iscnId := q.Get("iscn_id"); iscnId != "" {
		events = append(events, types.StringEvent{
			Type: "iscn_record",
			Attributes: []types.Attribute{
				{
					Key:   "iscn_id",
					Value: iscnId,
				},
			},
		})
		hasQuery = true
	}
	query := db.ISCNRecordQuery{}
	if fingerprint := q.Get("fingerprint"); fingerprint != "" {
		query.ContentFingerprints = []string{fingerprint}
		hasQuery = true
	}
	keywords := db.Keywords(q["keywords"])
	if len(keywords) > 0 {
		hasQuery = true
	}
	if sId, sName := q.Get("stakeholders.entity.id"), q.Get("stakeholders.entity.name"); sId != "" || sName != "" {
		query.Stakeholders = []db.Stakeholder{
			{
				Entity: &db.Entity{
					Id:   sId,
					Name: sName,
				},
			},
		}
		hasQuery = true
	}
	p := getPagination(q)

	pool := getPool(c)
	var records []iscnTypes.QueryResponseRecord
	var err error
	if hasQuery {
		records, err = db.QueryISCN(pool, events, query, keywords, p)
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
