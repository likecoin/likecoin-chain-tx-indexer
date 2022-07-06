package extractor

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

type nftClassMessage struct {
	Input db.NftClass `json:"input"`
}

func createNftClass(batch *db.Batch, messageData []byte, event types.StringEvents, timestamp time.Time) error {
	var message nftClassMessage
	if err := json.Unmarshal(messageData, &message); err != nil {
		panic(err)
	}
	var c db.NftClass = message.Input
	c.Id = utils.GetEventsValue(event, "likechain.likenft.EventNewClass", "class_id")
	c.Parent = getNftParent(event)
	sql := `
	INSERT INTO nft_class
	(id, parent_type, parent_iscn_id_prefix, parent_account,
	name, symbol, description, uri, uri_hash,
	metadata, config, price)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	batch.Batch.Queue(sql,
		c.Id, c.Parent.Type, c.Parent.IscnIdPrefix, c.Parent.Account,
		c.Name, c.Symbol, c.Description, c.URI, c.URIHash,
		c.Metadata, c.Config, c.Price,
	)
	return nil
}

func getNftParent(event types.StringEvents) db.NftClassParent {
	p := db.NftClassParent{
		IscnIdPrefix: utils.GetEventsValue(event, "likechain.likenft.EventNewClass", "parent_iscn_id_prefix"),
		Account:      utils.GetEventsValue(event, "likechain.likenft.EventNewClass", "parent_account"),
	}
	if p.IscnIdPrefix != "" {
		p.Type = "ISCN"
	} else if p.Account != "" {
		p.Type = "ACCOUNT"
	} else {
		p.Type = "UNKNOWN"
	}
	return p
}
