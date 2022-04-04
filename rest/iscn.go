package rest

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	iscnTypes "github.com/likecoin/likechain/x/iscn/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

type ISCNRecordsResponse struct {
	Records []iscnTypes.QueryResponseRecord `json:"records"`
}

func handleISCN(c *gin.Context) {
	q := c.Request.URL.Query()
	provided := false

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
		provided = true
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
		provided = true
	}
	query := db.ISCNRecordQuery{}
	if fingerprint := q.Get("fingerprint"); fingerprint != "" {
		query.ContentFingerprints = []string{fingerprint}
		provided = true
	}
	keywords := db.Keywords(q["keywords"])
	if len(keywords) > 0 {
		provided = true
	}
	p := getPagination(q)
	log.Println(query, events, p)
	conn := getConn(c)
	var records []iscnTypes.QueryResponseRecord
	var err error
	if provided {
		records, err = db.QueryISCN(conn, events, query, keywords, p)
	} else {
		records, err = db.QueryISCNList(conn, p)
	}
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	respondRecords(c, records)
}

func getPagination(q url.Values) db.Pagination {
	p := db.Pagination{
		Limit: 10,
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
