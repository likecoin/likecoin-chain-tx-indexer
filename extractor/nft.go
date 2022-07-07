package extractor

import (
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("Failed to unmarshal NFT class message: %w", err)
	}
	var c db.NftClass = message.Input
	c.Id = utils.GetEventsValue(event, "likechain.likenft.EventNewClass", "class_id")
	c.Parent = getNftParent(event)
	batch.InsertNftClass(c)
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

func mintNft(batch *db.Batch, messageData []byte, event types.StringEvents, timestamp time.Time) error {
	var message struct {
		Input db.Nft
	}
	if err := json.Unmarshal(messageData, &message); err != nil {
		return fmt.Errorf("Failed to unmarshal mint NFT message: %w", err)
	}
	nft := message.Input
	nft.NftId = utils.GetEventsValue(event, "likechain.likenft.EventMintNFT", "nft_id")
	nft.Owner = utils.GetEventsValue(event, "likechain.likenft.EventMintNFT", "owner")
	nft.ClassId = utils.GetEventsValue(event, "likechain.likenft.EventMintNFT", "class_id")
	batch.InsertNft(nft)
	return nil
}
