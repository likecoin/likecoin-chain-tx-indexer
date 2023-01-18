package extractor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

// TODO: handle Authz by extracting corresponding message from MsgExec sub-messages
func parseMessage(payload db.EventPayload) (db.NftMarketplaceItem, error) {
	var item struct {
		ClassId    string    `json:"class_id"`
		NftId      string    `json:"nft_id"`
		Creator    string    `json:"creator"`
		Price      string    `json:"price"`
		Expiration time.Time `json:"expiration"`
	}
	err := json.Unmarshal(payload.GetMessage(), &item)
	if err != nil {
		return db.NftMarketplaceItem{}, fmt.Errorf("failed to unmarshal marketplace related message: %w", err)
	}
	price := uint64(0)
	if item.Price != "" {
		price, err = strconv.ParseUint(item.Price, 10, 64)
		if err != nil {
			return db.NftMarketplaceItem{}, fmt.Errorf("failed to parse price in marketplace related message: %w", err)
		}
	}
	return db.NftMarketplaceItem{
		ClassId:    item.ClassId,
		NftId:      item.NftId,
		Creator:    item.Creator,
		Price:      price,
		Expiration: item.Expiration,
	}, nil
}

func createListing(payload db.EventPayload, event *types.StringEvent) error {
	item, err := parseMessage(payload)
	if err != nil {
		return err
	}
	item.Type = "listing"
	payload.Batch.InsertNFTMarketplaceItem(item)
	return nil
}

func deleteListing(payload db.EventPayload, event *types.StringEvent) error {
	item := db.NftMarketplaceItem{
		Type:    "listing",
		ClassId: utils.GetEventValue(event, "class_id"),
		NftId:   utils.GetEventValue(event, "nft_id"),
		Creator: utils.GetEventValue(event, "seller"),
	}
	payload.Batch.DeleteNFTMarketplaceItem(item)
	return nil
}

func updateListing(payload db.EventPayload, event *types.StringEvent) error {
	err := deleteListing(payload, event)
	if err != nil {
		return err
	}
	return createListing(payload, event)
}
func createOffer(payload db.EventPayload, event *types.StringEvent) error {
	item, err := parseMessage(payload)
	if err != nil {
		return err
	}
	item.Type = "offer"
	payload.Batch.InsertNFTMarketplaceItem(item)
	return nil
}

func deleteOffer(payload db.EventPayload, event *types.StringEvent) error {
	item := db.NftMarketplaceItem{
		Type:    "offer",
		ClassId: utils.GetEventValue(event, "class_id"),
		NftId:   utils.GetEventValue(event, "nft_id"),
		Creator: utils.GetEventValue(event, "buyer"),
	}
	payload.Batch.DeleteNFTMarketplaceItem(item)
	return nil
}

func updateOffer(payload db.EventPayload, event *types.StringEvent) error {
	err := deleteOffer(payload, event)
	if err != nil {
		return err
	}
	return createOffer(payload, event)
}

func getPriceFromEvent(event *types.StringEvent) uint64 {
	priceStr := utils.GetEventValue(event, "price")
	price, err := strconv.ParseUint(priceStr, 10, 64)
	if err != nil {
		// TODO: should we return error?
		return 0
	}
	return price
}

func marketplaceDeal(payload db.EventPayload, event *types.StringEvent) error {
	e := extractNftEvent(event, "class_id", "nft_id", "seller", "buyer")
	e.Price = getPriceFromEvent(event)
	sql := `UPDATE nft SET owner = $1 WHERE class_id = $2 AND nft_id = $3`
	payload.Batch.Batch.Queue(sql, e.Receiver, e.ClassId, e.NftId)
	e.Attach(payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}

func init() {
	eventExtractor.RegisterType("likechain.likenft.v1.EventBuyNFT", marketplaceDeal)
	eventExtractor.RegisterType("likechain.likenft.v1.EventSellNFT", marketplaceDeal)
	eventExtractor.RegisterType("likechain.likenft.v1.EventCreateListing", createListing)
	eventExtractor.RegisterType("likechain.likenft.v1.EventUpdateListing", updateListing)
	eventExtractor.RegisterType("likechain.likenft.v1.EventDeleteListing", deleteListing)
	eventExtractor.RegisterType("likechain.likenft.v1.EventCreateOffer", createOffer)
	eventExtractor.RegisterType("likechain.likenft.v1.EventUpdateOffer", updateOffer)
	eventExtractor.RegisterType("likechain.likenft.v1.EventDeleteOffer", deleteOffer)
}
