package extractor

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func attachNftEvent(e *db.NftEvent, payload *Payload) {
	events := payload.GetEvents()
	e.Events = events
	e.Timestamp = payload.Timestamp
	e.TxHash = payload.TxHash
	e.Memo = payload.Memo
	e.MsgIndex = payload.MsgIndex
}

type nftClassMessage struct {
	Input   db.NftClass `json:"input"`
	Creator string      `json:"creator"`
}

func createNftClass(payload *Payload, event *types.StringEvent) error {
	var message nftClassMessage
	if err := json.Unmarshal(payload.GetMessage(), &message); err != nil {
		return fmt.Errorf("failed to unmarshal NFT class message: %w", err)
	}
	c := message.Input
	c.Id = utils.GetEventValue(event, "class_id")
	c.Parent = getNftParent(event)
	c.CreatedAt = payload.Timestamp
	payload.Batch.InsertNftClass(c)

	e := db.NftEvent{
		ClassId: c.Id,
		Sender:  message.Creator,
		Action:  db.ACTION_NEW_CLASS,
	}
	attachNftEvent(&e, payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}

func updateNftClass(payload *Payload, event *types.StringEvent) error {
	var message nftClassMessage
	if err := json.Unmarshal(payload.GetMessage(), &message); err != nil {
		return fmt.Errorf("failed to unmarshal NFT class message: %w", err)
	}
	c := message.Input
	c.Id = utils.GetEventValue(event, "class_id")
	payload.Batch.UpdateNftClass(c)

	e := db.NftEvent{
		ClassId: c.Id,
		Sender:  message.Creator,
		Action:  db.ACTION_UPDATE_CLASS,
	}
	attachNftEvent(&e, payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}

func getNftParent(event *types.StringEvent) db.NftClassParent {
	p := db.NftClassParent{
		IscnIdPrefix: utils.GetEventValue(event, "parent_iscn_id_prefix"),
		Account:      utils.GetEventValue(event, "parent_account"),
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

func mintNft(payload *Payload, event *types.StringEvent) error {
	var message struct {
		Input db.Nft
	}
	if err := json.Unmarshal(payload.GetMessage(), &message); err != nil {
		return fmt.Errorf("failed to unmarshal mint NFT message: %w", err)
	}
	nft := message.Input
	nft.NftId = utils.GetEventValue(event, "nft_id")
	nft.Owner = utils.GetEventValue(event, "owner")
	nft.ClassId = utils.GetEventValue(event, "class_id")

	payload.Batch.InsertNft(nft)

	e := db.NftEvent{
		ClassId: nft.ClassId,
		NftId:   nft.NftId,
		Sender:  nft.Owner,
		Action:  db.ACTION_MINT,
	}
	attachNftEvent(&e, payload)
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
		logger.L.Warnw("Failed to parse price from event", "price_str", priceStr, "error", err)
		return 0
	}
	return coin.Amount.Uint64()
}

func extractNftEvent(event *types.StringEvent, classIdField, nftIdField, senderField, receiverField string) db.NftEvent {
	classId := utils.GetEventValue(event, classIdField)
	nftId := utils.GetEventValue(event, nftIdField)
	sender := utils.GetEventValue(event, senderField)
	receiver := utils.GetEventValue(event, receiverField)
	e := db.NftEvent{
		ClassId:  classId,
		NftId:    nftId,
		Sender:   sender,
		Receiver: receiver,
	}
	return e
}

func sendNft(payload *Payload, event *types.StringEvent) error {
	e := extractNftEvent(event, "class_id", "id", "sender", "receiver")
	e.Action = db.ACTION_SEND
	sql := `UPDATE nft SET owner = $1 WHERE class_id = $2 AND nft_id = $3`
	payload.Batch.Batch.Queue(sql, e.Receiver, e.ClassId, e.NftId)

	// In our application, we use authz token send together with NFT send to mimic
	// selling an NFT, where the API address is the "market" holding the NFT.
	// Buyer authorizes the API address ("market") to send LIKE to the market
	// or the seller, and the API address packs the send NFT message in the same
	// transaction, so the NFT send and token send are atomic.
	// We want to identify this case and extract the "price" from the transaction.

	// We assume the first message is the authz message with token send
	// TODO: also check if authz grantee is API address
	sendNftMsgIndex := payload.MsgIndex
	if sendNftMsgIndex > 0 {
		prevMsgEvents := payload.EventsList[sendNftMsgIndex-1].Events
		prevMsgAction := utils.GetEventsValue(prevMsgEvents, "message", "action")
		if prevMsgAction == "/cosmos.authz.v1beta1.MsgExec" {
			e.Price = extractPriceFromEvents(prevMsgEvents)

			incomes := GetIncomesFromSendNftMsgs(payload.EventsList, sendNftMsgIndex, payload.TxHash)
			for _, income := range incomes {
				payload.Batch.InsertNftIncome(income)
			}
		}
	}
	attachNftEvent(&e, payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}

func GetIncomesFromSendNftMsgs(eventsList db.EventsList, msgIndex int, txHash string) []db.NftIncome {
	if msgIndex < 1 {
		return []db.NftIncome{}
	}
	// We assume the first message is the authz message with token send
	authzMsgIndex := msgIndex - 1
	authzMsgEvents := eventsList[authzMsgIndex].Events
	authzMsgAction := utils.GetEventsValue(authzMsgEvents, "message", "action")
	if authzMsgAction != "/cosmos.authz.v1beta1.MsgExec" {
		return []db.NftIncome{}
	}

	sendTokenMsgStartIndex := msgIndex + 1
	rawIncomes := []utils.RawIncome{}
	for i := sendTokenMsgStartIndex; i < len(eventsList); i++ {
		currMsgEvents := eventsList[i].Events
		currMsgAction := utils.GetEventsValue(currMsgEvents, "message", "action")
		if currMsgAction != "/cosmos.bank.v1beta1.MsgSend" {
			break
		}
		address := utils.GetEventsValue(currMsgEvents, "coin_received", "receiver")
		amount := extractPriceFromEvents(currMsgEvents)
		rawIncomes = append(rawIncomes, utils.RawIncome{
			Address:   address,
			Amount:    amount,
			IsRoyalty: true, // first-hand selling revenue is classified as royalty
		})
	}

	aggregatedIncomes := utils.AggregateRawIncomes(rawIncomes)

	sendNftMsgEvnets := eventsList[msgIndex].Events
	classId := utils.GetEventsValue(sendNftMsgEvnets, "cosmos.nft.v1beta1.EventSend", "class_id")
	nftId := utils.GetEventsValue(sendNftMsgEvnets, "cosmos.nft.v1beta1.EventSend", "id")

	incomes := []db.NftIncome{}
	for _, income := range aggregatedIncomes {
		incomes = append(incomes, db.NftIncome{
			ClassId:   classId,
			NftId:     nftId,
			TxHash:    txHash,
			Address:   income.Address,
			Amount:    income.Amount,
			IsRoyalty: income.IsRoyalty,
		})
	}

	return incomes
}

func init() {
	eventExtractor.RegisterType("likechain.likenft.v1.EventNewClass", createNftClass)
	eventExtractor.RegisterType("likechain.likenft.v1.EventUpdateClass", updateNftClass)
	eventExtractor.RegisterType("likechain.likenft.v1.EventMintNFT", mintNft)
	eventExtractor.RegisterType("cosmos.nft.v1beta1.EventSend", sendNft)
}
