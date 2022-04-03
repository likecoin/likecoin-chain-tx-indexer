package rest

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	iscnTypes "github.com/likecoin/likechain/x/iscn/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

type ISCNRecordsResponse struct {
	Records []iscnTypes.QueryResponseRecord `json:"records"`
}

func handleISCNById(c *gin.Context) {
	q := c.Request.URL.Query()
	iscnId := q.Get("iscn_id")
	if iscnId == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "ISCN id not provided"})
		return
	}
	events := types.StringEvents{
		types.StringEvent{
			Type: "iscn_record",
			Attributes: []types.Attribute{
				{
					Key:   "iscn_id",
					Value: iscnId,
				},
			},
		},
	}
	conn := getConn(c)

	iscnInputs, err := db.QueryISCNByEvents(conn, events)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	respondRecords(c, iscnInputs)
}

func handleISCNByOwner(c *gin.Context) {
	q := c.Request.URL.Query()
	owner := q.Get("owner")
	if owner == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "owner is not provided"})
		return
	}
	log.Println(owner)
	events := types.StringEvents{
		types.StringEvent{
			Type: "iscn_record",
			Attributes: []types.Attribute{
				{
					Key:   "owner",
					Value: owner,
				},
			},
		},
	}
	conn := getConn(c)

	iscnInputs, err := db.QueryISCNByEvents(conn, events)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	respondRecords(c, iscnInputs)
}

func handleISCNByFingerprint(c *gin.Context) {
	q := c.Request.URL.Query()
	fingerprint := q.Get("fingerprint")
	if fingerprint == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "fingerprint not provided"})
		return
	}
	conn := getConn(c)

	query := fmt.Sprintf(`{"contentFingerprints": ["%s"]}`, fingerprint)

	iscnInputs, err := db.QueryISCNByRecord(conn, query)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	respondRecords(c, iscnInputs)
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
	log.Println(string(resJson))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(resJson)
}
