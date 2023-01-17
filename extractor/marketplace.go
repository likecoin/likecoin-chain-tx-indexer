package extractor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
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

func createListing(payload db.EventPayload) error {
	item, err := parseMessage(payload)
	if err != nil {
		return err
	}
	item.Type = "listing"
	payload.Batch.InsertNFTMarketplaceItem(item)
	return nil
}

func deleteListing(payload db.EventPayload) error {
	item, err := parseMessage(payload)
	if err != nil {
		return err
	}
	item.Type = "listing"
	payload.Batch.DeleteNFTMarketplaceItem(item)
	return nil
}

func updateListing(payload db.EventPayload) error {
	err := deleteListing(payload)
	if err != nil {
		return err
	}
	return createListing(payload)
}
func createOffer(payload db.EventPayload) error {
	item, err := parseMessage(payload)
	if err != nil {
		return err
	}
	item.Type = "offer"
	payload.Batch.InsertNFTMarketplaceItem(item)
	return nil
}

func deleteOffer(payload db.EventPayload) error {
	item, err := parseMessage(payload)
	if err != nil {
		return err
	}
	item.Type = "offer"
	payload.Batch.DeleteNFTMarketplaceItem(item)
	return nil
}

func updateOffer(payload db.EventPayload) error {
	err := deleteOffer(payload)
	if err != nil {
		return err
	}
	return createOffer(payload)
}

func buyNft(payload db.EventPayload) error {
	e := extractNftEvent(payload.GetEvents(), "likechain.likenft.v1.EventBuyNFT", "class_id", "nft_id", "seller", "buyer")
	payload.Batch.DeleteNFTMarketplaceItem(db.NftMarketplaceItem{
		Type:    "listing",
		ClassId: e.ClassId,
		NftId:   e.NftId,
		Creator: e.Sender,
	})
	transferNftOwnershipFromEvent(payload, e)
	return nil
}

func sellNft(payload db.EventPayload) error {
	e := extractNftEvent(payload.GetEvents(), "likechain.likenft.v1.EventSellNFT", "class_id", "nft_id", "seller", "buyer")
	payload.Batch.DeleteNFTMarketplaceItem(db.NftMarketplaceItem{
		Type:    "offer",
		ClassId: e.ClassId,
		NftId:   e.NftId,
		Creator: e.Receiver,
	})
	transferNftOwnershipFromEvent(payload, e)
	return nil
}

func init() {
	eventExtractor.Register("message", "action", "buy_nft", buyNft)
	eventExtractor.Register("message", "action", "sell_nft", sellNft)
	eventExtractor.Register("message", "action", "create_listing", createListing)
	eventExtractor.Register("message", "action", "update_listing", updateListing)
	eventExtractor.Register("message", "action", "delete_listing", deleteListing)
	eventExtractor.Register("message", "action", "create_offer", createOffer)
	eventExtractor.Register("message", "action", "update_offer", updateOffer)
	eventExtractor.Register("message", "action", "delete_offer", deleteOffer)
}
