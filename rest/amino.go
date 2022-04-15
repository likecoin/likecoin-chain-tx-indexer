package rest

import (
	"fmt"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func convertToStdTx(txBytes []byte) (legacytx.StdTx, error) {
	txI, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return legacytx.StdTx{}, err
	}

	tx, ok := txI.(signing.Tx)
	if !ok {
		return legacytx.StdTx{}, fmt.Errorf("%+v is not backwards compatible with %T", tx, legacytx.StdTx{})
	}

	return clienttx.ConvertTxToStdTx(encodingConfig.Amino, tx)
}

func packStdTxResponse(txRes *types.TxResponse) error {
	txBytes := txRes.Tx.Value
	stdTx, err := convertToStdTx(txBytes)
	if err != nil {
		return err
	}

	// Pack the amino stdTx into the TxResponse's Any.
	txRes.Tx = codectypes.UnsafePackAny(stdTx)
	return nil
}

func handleAminoTxsSearch(c *gin.Context) {
	q := c.Request.URL.Query()
	page, err := getPage(q, "page")
	height := uint64(0)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	limit, err := getLimit(q, "limit")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	events, err := getEvents(q)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	conn := getConn(c)
	totalCount, err := db.QueryCount(conn, events, height)
	if err != nil {
		logger.L.Errorw("Cannot get total tx count from database", "events", events, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	maxPage := (totalCount-1)/limit + 1
	if maxPage == 0 {
		maxPage = 1
	}
	if page > maxPage {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("page should be within [1, %d] range, given %d", maxPage, page)})
		return
	}
	offset := limit * (page - 1)
	txs, err := db.QueryTxs(conn, events, height, limit, offset, db.ORDER_ASC)
	if err != nil {
		logger.L.Errorw("Cannot get txs from database", "events", events, "limit", limit, "page", page, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	searchTxsResult := types.NewSearchTxsResult(totalCount, uint64(len(txs)), page, limit, txs)

	for _, txRes := range searchTxsResult.Txs {
		packStdTxResponse(txRes)
	}

	json, err := encodingConfig.Amino.MarshalJSON(searchTxsResult)
	if err != nil {
		logger.L.Errorw("Cannot convert searchTxsResult to JSON", "events", events, "limit", limit, "page", page, "error", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(200)
	c.Writer.Write(json)
}
