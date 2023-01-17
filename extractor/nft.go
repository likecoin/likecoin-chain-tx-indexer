package extractor

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

type nftClassMessage struct {
	Input   db.NftClass `json:"input"`
	Creator string      `json:"creator"`
}

func createNftClass(payload db.EventPayload) error {
	var message nftClassMessage
	if err := json.Unmarshal(payload.GetMessage(), &message); err != nil {
		return fmt.Errorf("failed to unmarshal NFT class message: %w", err)
	}
	c := message.Input
	events := payload.GetEvents()
	c.Id = utils.GetEventsValue(events, "likechain.likenft.v1.EventNewClass", "class_id")
	c.Parent = getNftParent(events, "likechain.likenft.v1.EventNewClass")
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
	if err := json.Unmarshal(payload.GetMessage(), &message); err != nil {
		return fmt.Errorf("failed to unmarshal NFT class message: %w", err)
	}
	c := message.Input
	events := payload.GetEvents()
	c.Id = utils.GetEventsValue(events, "likechain.likenft.v1.EventUpdateClass", "class_id")
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
	if err := json.Unmarshal(payload.GetMessage(), &message); err != nil {
		return fmt.Errorf("failed to unmarshal mint NFT message: %w", err)
	}
	events := payload.GetEvents()
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

func extractPriceFromEvents(events types.StringEvents) uint64 {
	priceStr := utils.GetEventsValue(events, "coin_received", "amount")
	if priceStr == "" {
		return 0
	}
	coin, err := types.ParseCoinNormalized(priceStr)
	if err != nil {
		logger.L.Warnw("Failed to parse price from event", "price_str", priceStr)
		return 0
	}
	return coin.Amount.Uint64()
}

func extractNftEvent(events types.StringEvents, typeField, classIdField, nftIdField, senderField, receiverField string) db.NftEvent {
	classId := utils.GetEventsValue(events, typeField, classIdField)
	nftId := utils.GetEventsValue(events, typeField, nftIdField)
	sender := utils.GetEventsValue(events, typeField, senderField)
	receiver := utils.GetEventsValue(events, typeField, receiverField)
	e := db.NftEvent{
		ClassId:  classId,
		NftId:    nftId,
		Sender:   sender,
		Receiver: receiver,
		Price:    extractPriceFromEvents(events),
	}
	return e
}

func transferNftOwnershipFromEvent(payload db.EventPayload, e db.NftEvent) {
	sql := `UPDATE nft SET owner = $1 WHERE class_id = $2 AND nft_id = $3`
	payload.Batch.Batch.Queue(sql, e.Receiver, e.ClassId, e.NftId)
	e.Attach(payload)
	payload.Batch.InsertNftEvent(e)
}

func sendNft(payload db.EventPayload) error {
	e := extractNftEvent(payload.GetEvents(), "cosmos.nft.v1beta1.EventSend", "class_id", "id", "sender", "receiver")

	// In our application, we use authz token send together with NFT send to mimic
	// selling an NFT, where the API address is the "market" holding the NFT.
	// Buyer authorizes the API address ("market") to send LIKE to the market
	// or the seller, and the API address packs the send NFT message in the same
	// transaction, so the NFT send and token send are atomic.
	// We want to identify this case and extract the "price" from the transaction.

	// We assume the first message is the authz message with token send
	// TODO: all price extraction should follow API address set in CLI / config
	events := payload.EventsList[0].Events
	action := utils.GetEventsValue(events, "message", "action")
	if action == "/cosmos.authz.v1beta1.MsgExec" {
		e.Price = extractPriceFromEvents(events)
	}
	transferNftOwnershipFromEvent(payload, e)
	return nil
}

func init() {
	eventExtractor.Register("message", "action", "new_class", createNftClass)
	eventExtractor.Register("message", "action", "update_class", updateNftClass)
	eventExtractor.Register("message", "action", "mint_nft", mintNft)
	eventExtractor.Register("message", "action", "/cosmos.nft.v1beta1.MsgSend", sendNft)
}
