package rest

import (
	"log"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	iscnTypes "github.com/likecoin/likechain/x/iscn/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleISCNById(c *gin.Context) {
	pool := getDB(c)
	q := c.Request.URL.Query()
	iscnId := q.Get("iscn_id")
	if iscnId == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "ISCN id not provided"})
		return
	}
	log.Println(iscnId)
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
	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	defer conn.Release()

	iscnInputs, err := db.QueryISCN(conn, events)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	if len(iscnInputs) == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "Record not found"})
		return
	}

	response := iscnTypes.QueryRecordsByIdResponse{
		Owner:         "",
		LatestVersion: 3,
		Records:       iscnInputs,
	}
	resJson, err := encodingConfig.Marshaler.MarshalJSON(&response)
	log.Println(string(resJson))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(resJson)
}

func handleISCNByOwner(c *gin.Context) {
	pool := getDB(c)
	q := c.Request.URL.Query()
	owner := q.Get("owner")
	if owner == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "ISCN id not provided"})
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
	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	defer conn.Release()

	iscnInputs, err := db.QueryISCN(conn, events)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	if len(iscnInputs) == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "Record not found"})
		return
	}

	response := iscnTypes.QueryRecordsByOwnerResponse{
		Records: iscnInputs,
	}
	resJson, err := encodingConfig.Marshaler.MarshalJSON(&response)
	log.Println(string(resJson))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(resJson)
}
