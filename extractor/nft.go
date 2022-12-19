package extractor

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

type nftClassMessage struct {
	Input   db.NftClass `json:"input"`
	Creator string      `json:"creator"`
}

func createNftClass(payload db.EventPayload) error {
	var message nftClassMessage
	if err := json.Unmarshal(payload.Message, &message); err != nil {
		return fmt.Errorf("failed to unmarshal NFT class message: %w", err)
	}
	var c db.NftClass = message.Input
	c.Id = utils.GetEventsValue(payload.Events, "likechain.likenft.v1.EventNewClass", "class_id")
	c.Parent = getNftParent(payload.Events, "likechain.likenft.v1.EventNewClass")
	c.CreatedAt = payload.Timestamp
	payload.Batch.InsertNftClass(c)

	e := db.NftEvent{
		ClassId: c.Id,
		Sender:  message.Creator,
	}
	e.Attach(payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}

func updateNftClass(payload db.EventPayload) error {
	var message nftClassMessage
	if err := json.Unmarshal(payload.Message, &message); err != nil {
		return fmt.Errorf("failed to unmarshal NFT class message: %w", err)
	}
	var c db.NftClass = message.Input
	c.Id = utils.GetEventsValue(payload.Events, "likechain.likenft.v1.EventUpdateClass", "class_id")
	payload.Batch.UpdateNftClass(c)

	e := db.NftEvent{
		ClassId: c.Id,
		Sender:  message.Creator,
	}
	e.Attach(payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}

func getNftParent(event types.StringEvents, eventType string) db.NftClassParent {
	p := db.NftClassParent{
		IscnIdPrefix: utils.GetEventsValue(event, eventType, "parent_iscn_id_prefix"),
		Account:      utils.GetEventsValue(event, eventType, "parent_account"),
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

func mintNft(payload db.EventPayload) error {
	var message struct {
		Input db.Nft
	}
	if err := json.Unmarshal(payload.Message, &message); err != nil {
		return fmt.Errorf("failed to unmarshal mint NFT message: %w", err)
	}
	events := payload.Events
	nft := message.Input
	nft.NftId = utils.GetEventsValue(events, "likechain.likenft.v1.EventMintNFT", "nft_id")
	nft.Owner = utils.GetEventsValue(events, "likechain.likenft.v1.EventMintNFT", "owner")
	nft.ClassId = utils.GetEventsValue(events, "likechain.likenft.v1.EventMintNFT", "class_id")
	payload.Batch.InsertNft(nft)

	e := db.NftEvent{
		ClassId: nft.ClassId,
		NftId:   nft.NftId,
		Sender:  nft.Owner,
	}
	e.Attach(payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}

func extractNftEvent(events types.StringEvents, typeField, classIdField, nftIdField, senderField, receiverField string) db.NftEvent {
	classId := utils.GetEventsValue(events, typeField, classIdField)
	nftId := utils.GetEventsValue(events, typeField, nftIdField)
	sender := utils.GetEventsValue(events, typeField, senderField)
	receiver := utils.GetEventsValue(events, typeField, receiverField)
	return db.NftEvent{
		ClassId:  classId,
		NftId:    nftId,
		Sender:   sender,
		Receiver: receiver,
	}
}

func defineNftChangeOwnerEventHandler(typeField, classIdField, nftIdField, senderField, receiverField string) db.EventHandler {
	return func(payload db.EventPayload) error {
		e := extractNftEvent(payload.Events, typeField, classIdField, nftIdField, senderField, receiverField)
		sql := `UPDATE nft SET owner = $1 WHERE class_id = $2 AND nft_id = $3`
		payload.Batch.Batch.Queue(sql, e.Receiver, e.ClassId, e.NftId)
		e.Attach(payload)
		payload.Batch.InsertNftEvent(e)
		return nil
	}
}

var sendNft = defineNftChangeOwnerEventHandler("cosmos.nft.v1beta1.EventSend", "class_id", "id", "sender", "receiver")
var buyNft = defineNftChangeOwnerEventHandler("likechain.likenft.v1.EventBuyNFT", "class_id", "nft_id", "seller", "buyer")
var sellNft = defineNftChangeOwnerEventHandler("likechain.likenft.v1.EventSellNFT", "class_id", "nft_id", "seller", "buyer")
