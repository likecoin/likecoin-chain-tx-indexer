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
	records := make([]iscnTypes.QueryResponseRecord, 0)
	for _, v := range iscnInputs {
		records = append(records, iscnTypes.QueryResponseRecord{
			Data: v,
		})
	}
	response := iscnTypes.QueryRecordsByIdResponse{
		Owner:         "",
		LatestVersion: 3,
		Records:       records,
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
