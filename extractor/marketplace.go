package extractor

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func parseMessage(payload *Payload) (db.NftMarketplaceItem, error) {
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
		// HACK: pg cannot store MaxUInt64
		if price > math.MaxInt64 {
			price = math.MaxInt64
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

func createListing(payload *Payload, event *types.StringEvent) error {
	item, err := parseMessage(payload)
	if err != nil {
		return err
	}
	item.Type = "listing"
	payload.Batch.DeleteNFTMarketplaceItemSilently(item)
	payload.Batch.InsertNFTMarketplaceItem(item)
	return nil
}

func deleteListing(payload *Payload, event *types.StringEvent) error {
	item := db.NftMarketplaceItem{
		Type:    "listing",
		ClassId: utils.GetEventValue(event, "class_id"),
		NftId:   utils.GetEventValue(event, "nft_id"),
		Creator: utils.GetEventValue(event, "seller"),
	}
	payload.Batch.DeleteNFTMarketplaceItem(item)
	return nil
}

func updateListing(payload *Payload, event *types.StringEvent) error {
	err := deleteListing(payload, event)
	if err != nil {
		return err
	}
	return createListing(payload, event)
}
func createOffer(payload *Payload, event *types.StringEvent) error {
	item, err := parseMessage(payload)
	if err != nil {
		return err
	}
	item.Type = "offer"
	payload.Batch.InsertNFTMarketplaceItem(item)
	return nil
}

func deleteOffer(payload *Payload, event *types.StringEvent) error {
	item := db.NftMarketplaceItem{
		Type:    "offer",
		ClassId: utils.GetEventValue(event, "class_id"),
		NftId:   utils.GetEventValue(event, "nft_id"),
		Creator: utils.GetEventValue(event, "buyer"),
	}
	payload.Batch.DeleteNFTMarketplaceItem(item)
	return nil
}

func updateOffer(payload *Payload, event *types.StringEvent) error {
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
	// HACK: pg cannot store MaxUInt64
	if price > math.MaxInt64 {
		price = math.MaxInt64
	}
	return price
}

func buyNft(payload *Payload, event *types.StringEvent) error {
	return marketplaceDeal(payload, event, db.ACTION_BUY)
}

func sellNft(payload *Payload, event *types.StringEvent) error {
	return marketplaceDeal(payload, event, db.ACTION_SELL)
}

func marketplaceDeal(payload *Payload, event *types.StringEvent, actionType db.NftEventAction) error {
	e := extractNftEvent(event, "class_id", "nft_id", "seller", "buyer")
	e.Price = getPriceFromEvent(event)
	e.Action = actionType
	sql := `UPDATE nft SET owner = $1 WHERE class_id = $2 AND nft_id = $3`
	payload.Batch.Batch.Queue(sql, e.Receiver, e.ClassId, e.NftId)

	msgIndex := payload.MsgIndex
	msgEvents := payload.EventsList[msgIndex].Events
	incomes := GetIncomesFromBuySellNftMsg(msgEvents, payload.TxHash)
	for _, income := range incomes {
		payload.Batch.InsertNftIncome(income)
	}
	attachNftEvent(&e, payload)
	payload.Batch.InsertNftEvent(e)
	return nil
}

// marketplace messages may contain multiple coin_received events,
// should not directly use GetEventValue() or GetEventsValue() since it only returns the first one
func GetIncomesFromBuySellNftMsg(events types.StringEvents, txHash string) []db.NftIncome {
	seller := ""
	for _, event := range events {
		seller = utils.GetEventValue(&event, "seller")
		if seller != "" {
			break
		}
	}

	rawIncomes := []utils.RawIncome{}
	address := ""
	amount := uint64(0)
	for _, event := range events {
		if event.Type == "coin_received" {
			for _, attr := range event.Attributes {
				if attr.Key == "receiver" {
					address = attr.Value
				}
				if attr.Key == "amount" {
					amountStr := attr.Value
					coin, err := types.ParseCoinNormalized(amountStr)
					if err != nil {
						logger.L.Warnw("Failed to parse income from event", "income_str", amountStr, "error", err)
						address = ""
						continue
					}
					amount = coin.Amount.Uint64()
				}
				if address != "" && amount != 0 {
					rawIncomes = append(rawIncomes, utils.RawIncome{
						Address:   address,
						Amount:    amount,
						IsRoyalty: address != seller,
					})
					address = ""
					amount = 0
				}
			}
		}
	}

	aggregatedIncomes := utils.AggregateRawIncomes(rawIncomes)

	classId := ""
	nftId := ""
	for _, event := range events {
		classId = utils.GetEventValue(&event, "class_id")
		nftId = utils.GetEventValue(&event, "nft_id")
		if classId != "" && nftId != "" {
			break
		}
	}

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
	eventExtractor.RegisterType("likechain.likenft.v1.EventBuyNFT", buyNft)
	eventExtractor.RegisterType("likechain.likenft.v1.EventSellNFT", sellNft)
	eventExtractor.RegisterType("likechain.likenft.v1.EventCreateListing", createListing)
	eventExtractor.RegisterType("likechain.likenft.v1.EventUpdateListing", updateListing)
	eventExtractor.RegisterType("likechain.likenft.v1.EventDeleteListing", deleteListing)
	eventExtractor.RegisterType("likechain.likenft.v1.EventCreateOffer", createOffer)
	eventExtractor.RegisterType("likechain.likenft.v1.EventUpdateOffer", updateOffer)
	eventExtractor.RegisterType("likechain.likenft.v1.EventDeleteOffer", deleteOffer)
}
