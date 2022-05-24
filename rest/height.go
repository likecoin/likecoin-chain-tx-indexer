package rest

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleLatestHeight(c *gin.Context) {
	conn := getConn(c)

	latestHeight, err := db.GetLatestHeight(conn)
	if err != nil {
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
					return
	}

	respondLatestHeight(c, latestHeight)
}

func respondLatestHeight(c *gin.Context, latestHeight int64) {

	type BlockState struct {
					LatestHeight int64 `json:"latest_height"`
	}

	response := BlockState{
					LatestHeight: latestHeight,
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