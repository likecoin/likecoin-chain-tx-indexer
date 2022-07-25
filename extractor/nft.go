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
	if err := json.Unmarshal(payload.Message, &message); err != nil {
		return fmt.Errorf("Failed to unmarshal NFT class message: %w", err)
	}
	var c db.NftClass = message.Input
	c.Id = utils.GetEventsValue(payload.Events, "likechain.likenft.v1.EventNewClass", "class_id")
	c.Parent = getNftParent(payload.Events, "likechain.likenft.v1.EventNewClass")
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
		return fmt.Errorf("Failed to unmarshal NFT class message: %w", err)
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
		return fmt.Errorf("Failed to unmarshal mint NFT message: %w", err)
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

func sendNft(payload db.EventPayload) error {
	events := payload.Events
	classId := utils.GetEventsValue(events, "cosmos.nft.v1beta1.EventSend", "class_id")
	nftId := utils.GetEventsValue(events, "cosmos.nft.v1beta1.EventSend", "id")
	sender := utils.GetEventsValue(events, "cosmos.nft.v1beta1.EventSend", "sender")
	receiver := utils.GetEventsValue(events, "cosmos.nft.v1beta1.EventSend", "receiver")
	sql := `UPDATE nft SET owner = $1 WHERE class_id = $2 AND nft_id = $3`
	payload.Batch.Batch.Queue(sql, receiver, classId, nftId)
	logger.L.Debugf("transfer nft %s from %s to %s\n", nftId, sender, receiver)

	e := db.NftEvent{
		ClassId:  classId,
		NftId:    nftId,
		Sender:   sender,
		Receiver: receiver,
	}
	e.Attach(payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}
