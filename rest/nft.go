package rest

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func handleNftByIscn(c *gin.Context) {
	q := c.Request.URL.Query()
	iscn := q.Get("iscn_id_prefix")
	if iscn == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "iscn_id_prefix is require"})
		return
	}

	conn := getConn(c)
	res, err := db.GetNftByIscn(conn, iscn)
	if err != nil {
		panic(err)
	}

	resJson, err := json.Marshal(&res)
	if err != nil {
		panic(err)
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(resJson)
}
