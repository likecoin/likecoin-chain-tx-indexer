package extractor_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/extractor"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestBuyNftOwnership(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_01_LIKE,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-102993",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_LIKE,
		},
	}
	txs := []string{
		fmt.Sprintf(`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgBuyNFT","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","seller":"%[4]s","price":"100000000000"}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"likechain.likenft.v1.EventBuyNFT","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[4]s\""},{"key":"buyer","value":"\"%[1]s\""},{"key":"price","value":"\"100000000000\""}]},{"type":"message","attributes":[{"key":"action","value":"buy_nft"},{"key":"sender","value":"%[1]s"}]}]}]}`, ADDR_02_LIKE, nftClasses[0].Id, nfts[0].NftId, ADDR_01_LIKE),
	}
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		Txs:        txs,
	})

	ownersRes, err := GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, ownersRes.Owners[0].Owner, ADDR_01_LIKE)

	eventRes, err := GetNftEvents(Conn, QueryEventsRequest{
		ClassId: nftClasses[0].Id,
	}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, eventRes.Events)

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, ownersRes.Owners[0].Owner, ADDR_02_LIKE)

	eventRes, err = GetNftEvents(Conn, QueryEventsRequest{
		ClassId: nftClasses[0].Id,
	}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, eventRes.Events, 1)
	require.Equal(t, eventRes.Events[0].Action, ACTION_BUY)
	require.Equal(t, eventRes.Events[0].NftId, nfts[0].NftId)
	require.Equal(t, eventRes.Events[0].Sender, ADDR_01_LIKE)
	require.Equal(t, eventRes.Events[0].Receiver, ADDR_02_LIKE)
	require.Equal(t, eventRes.Events[0].TxHash, "AAAAAA")
}

func TestSellNftOwnership(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_01_LIKE,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-919775",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_LIKE,
		},
	}
	txs := []string{
		fmt.Sprintf(`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgSellNFT","creator":"%[4]s","class_id":"%[2]s","nft_id":"%[3]s","buyer":"%[1]s","price":"100000000000"}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"likechain.likenft.v1.EventSellNFT","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[4]s\""},{"key":"buyer","value":"\"%[1]s\""},{"key":"price","value":"\"100000000000\""}]},{"type":"message","attributes":[{"key":"action","value":"sell_nft"},{"key":"sender","value":"%[4]s"}]}]}]}`, ADDR_02_LIKE, nftClasses[0].Id, nfts[0].NftId, ADDR_01_LIKE),
	}
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		Txs:        txs,
	})

	ownersRes, err := GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, ownersRes.Owners[0].Owner, ADDR_01_LIKE)

	eventRes, err := GetNftEvents(Conn, QueryEventsRequest{
		ClassId: nftClasses[0].Id,
	}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, eventRes.Events)

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, ownersRes.Owners[0].Owner, ADDR_02_LIKE)

	eventRes, err = GetNftEvents(Conn, QueryEventsRequest{
		ClassId: nftClasses[0].Id,
	}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, eventRes.Events, 1)
	require.Equal(t, eventRes.Events[0].Action, ACTION_SELL)
	require.Equal(t, eventRes.Events[0].NftId, nfts[0].NftId)
	require.Equal(t, eventRes.Events[0].Sender, ADDR_01_LIKE)
	require.Equal(t, eventRes.Events[0].Receiver, ADDR_02_LIKE)
	require.Equal(t, eventRes.Events[0].TxHash, "AAAAAA")
}

func TestListing(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_01_LIKE,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-100023",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-100024",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_02_LIKE,
		},
	}
	initPrice1 := uint64(100000000000)
	initPrice2 := uint64(100000000001)
	expiration := time.Unix(1700000000, 0).UTC()
	txs := []string{
		fmt.Sprintf(
			`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgCreateListing","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","price":"%[4]d","expiration":"%[5]s"}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"create_listing"}]},{"type":"likechain.likenft.v1.EventCreateListing","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[1]s\""}]}]}]}`,
			ADDR_01_LIKE, nftClasses[0].Id, nfts[0].NftId, initPrice1, expiration.Format(time.RFC3339),
		),
		fmt.Sprintf(
			`{"txhash":"AAAAAB","height":"1235","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgCreateListing","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","price":"%[4]d","expiration":"%[5]s"}],"memo":"AAAAAB"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"create_listing"}]},{"type":"likechain.likenft.v1.EventCreateListing","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[1]s\""}]}]}]}`,
			ADDR_02_LIKE, nftClasses[0].Id, nfts[1].NftId, initPrice2, expiration.Add(1*time.Second).Format(time.RFC3339),
		),
	}
	blockTime := expiration.Add(-10000 * time.Second)
	InsertTestData(DBTestData{
		Iscns:           iscns,
		NftClasses:      nftClasses,
		Nfts:            nfts,
		Txs:             txs,
		LatestBlockTime: &blockTime,
	})

	ownersRes, err := GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 2)

	itemsRes, err := GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "listing"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, itemsRes.Items)

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 2)

	itemsRes, err = GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "listing"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, itemsRes.Items, 2)
	require.Equal(t, "listing", itemsRes.Items[0].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[0].ClassId)
	require.Equal(t, nfts[0].NftId, itemsRes.Items[0].NftId)
	require.Equal(t, ADDR_01_LIKE, itemsRes.Items[0].Creator)
	require.Equal(t, initPrice1, itemsRes.Items[0].Price)
	require.Equal(t, expiration, itemsRes.Items[0].Expiration)
	require.Equal(t, "listing", itemsRes.Items[1].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[1].ClassId)
	require.Equal(t, nfts[1].NftId, itemsRes.Items[1].NftId)
	require.Equal(t, ADDR_02_LIKE, itemsRes.Items[1].Creator)
	require.Equal(t, initPrice2, itemsRes.Items[1].Price)
	require.Equal(t, expiration.Add(1*time.Second), itemsRes.Items[1].Expiration)

	updatedPrice1 := uint64(100000000002)
	txs = []string{
		fmt.Sprintf(
			`{"txhash":"AAAAAC","height":"1236","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgUpdateListing","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","price":"%[4]d","expiration":"%[5]s"}],"memo":"AAAAAC"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"update_listing"}]},{"type":"likechain.likenft.v1.EventUpdateListing","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[1]s\""}]}]}]}`,
			ADDR_01_LIKE, nftClasses[0].Id, nfts[0].NftId, updatedPrice1, expiration.Add(2*time.Second).Format(time.RFC3339),
		),
	}
	InsertTestData(DBTestData{Txs: txs})

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 2)

	itemsRes, err = GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "listing"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, itemsRes.Items, 2)
	require.Equal(t, "listing", itemsRes.Items[0].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[0].ClassId)
	require.Equal(t, nfts[1].NftId, itemsRes.Items[0].NftId)
	require.Equal(t, ADDR_02_LIKE, itemsRes.Items[0].Creator)
	require.Equal(t, initPrice2, itemsRes.Items[0].Price)
	require.Equal(t, expiration.Add(1*time.Second), itemsRes.Items[0].Expiration)
	require.Equal(t, "listing", itemsRes.Items[1].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[1].ClassId)
	require.Equal(t, nfts[0].NftId, itemsRes.Items[1].NftId)
	require.Equal(t, ADDR_01_LIKE, itemsRes.Items[1].Creator)
	require.Equal(t, updatedPrice1, itemsRes.Items[1].Price)
	require.Equal(t, expiration.Add(2*time.Second), itemsRes.Items[1].Expiration)

	txs = []string{
		fmt.Sprintf(
			`{"txhash":"AAAAAD","height":"1237","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgDeleteListing","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s"}],"memo":"AAAAAD"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"delete_listing"}]},{"type":"likechain.likenft.v1.EventDeleteListing","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[1]s\""}]}]}]}`,
			ADDR_02_LIKE, nftClasses[0].Id, nfts[1].NftId,
		),
	}
	InsertTestData(DBTestData{Txs: txs})

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 2)

	itemsRes, err = GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "listing"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, itemsRes.Items, 1)
	require.Equal(t, "listing", itemsRes.Items[0].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[0].ClassId)
	require.Equal(t, nfts[0].NftId, itemsRes.Items[0].NftId)
	require.Equal(t, ADDR_01_LIKE, itemsRes.Items[0].Creator)
	require.Equal(t, updatedPrice1, itemsRes.Items[0].Price)
	require.Equal(t, expiration.Add(2*time.Second), itemsRes.Items[0].Expiration)

	stakeholder1 := ADDR_03_LIKE
	stakeholder2 := ADDR_01_LIKE
	royalty1 := uint64(10000000000)
	royalty2 := updatedPrice1 - royalty1

	txs = []string{
		fmt.Sprintf(
			`{"txhash":"AAAAAE","height":"1238","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgBuyNFT","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","seller":"%[4]s","price":"%[5]d"}],"memo":"AAAAAE"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[6]s"},{"key":"amount","value":"%[8]dnanolike"},{"key":"receiver","value":"%[7]s"},{"key":"amount","value":"%[9]dnanolike"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[1]s"},{"key":"amount","value":"%[8]dnanolike"},{"key":"spender","value":"%[1]s"},{"key":"amount","value":"%[9]dnanolike"}]},{"type":"likechain.likenft.v1.EventBuyNFT","attributes":[{"key":"buyer","value":"\"%[1]s\""},{"key":"price","value":"\"%[5]d\""},{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[4]s\""}]},{"type":"likechain.likenft.v1.EventDeleteListing","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[4]s\""}]},{"type":"message","attributes":[{"key":"action","value":"buy_nft"},{"key":"sender","value":"%[1]s"},{"key":"sender","value":"%[1]s"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[6]s"},{"key":"sender","value":"%[1]s"},{"key":"amount","value":"%[8]dnanolike"},{"key":"recipient","value":"%[7]s"},{"key":"sender","value":"%[1]s"},{"key":"amount","value":"%[9]dnanolike"}]}]}]}`,
			ADDR_02_LIKE, nftClasses[0].Id, nfts[0].NftId, ADDR_01_LIKE, updatedPrice1, stakeholder1, stakeholder2, royalty1, royalty2,
		),
	}
	InsertTestData(DBTestData{Txs: txs})

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, ADDR_02_LIKE, ownersRes.Owners[0].Owner)
	require.Len(t, ownersRes.Owners[0].Nfts, 2)

	itemsRes, err = GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "listing"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, itemsRes.Items)

	eventsRes, err := GetNftEvents(Conn,
		QueryEventsRequest{
			ClassId:    nftClasses[0].Id,
			NftId:      nfts[0].NftId,
			ActionType: []NftEventAction{ACTION_BUY},
		},
		PageRequest{Limit: 1, Reverse: true},
	)
	require.NoError(t, err)
	require.Len(t, eventsRes.Events, 1)
	require.Equal(t, updatedPrice1, eventsRes.Events[0].Price)

	royalties, err := GetNftRoyalties(Conn,
		QueryRoyaltiesRequest{
			ClassId: nftClasses[0].Id,
			NftId:   nfts[0].NftId,
			OrderBy: "royalty",
		}, PageRequest{Limit: 10},
	)
	require.NoError(t, err)
	require.Len(t, royalties.Royalties, 2)
	require.Equal(t, nftClasses[0].Id, royalties.Royalties[0].ClassId)
	require.Equal(t, nfts[0].NftId, royalties.Royalties[0].NftId)
	require.Equal(t, "AAAAAE", royalties.Royalties[0].TxHash)
	require.Equal(t, stakeholder1, royalties.Royalties[0].Stakeholder)
	require.Equal(t, royalty1, royalties.Royalties[0].Royalty)
	require.Equal(t, stakeholder2, royalties.Royalties[1].Stakeholder)
	require.Equal(t, royalty2, royalties.Royalties[1].Royalty)
}

func TestOffer(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_01_LIKE,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-58177",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-58178",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_02_LIKE,
		},
	}
	initPrice1 := uint64(100000000000)
	initPrice2 := uint64(100000000001)
	expiration := time.Unix(1700000000, 0).UTC()
	txs := []string{
		fmt.Sprintf(
			`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgCreateOffer","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","price":"%[4]d","expiration":"%[5]s"}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"create_offer"}]},{"type":"likechain.likenft.v1.EventCreateOffer","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"buyer","value":"\"%[1]s\""}]}]}]}`,
			ADDR_02_LIKE, nftClasses[0].Id, nfts[0].NftId, initPrice1, expiration.Format(time.RFC3339),
		),
		fmt.Sprintf(
			`{"txhash":"AAAAAB","height":"1235","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgCreateOffer","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","price":"%[4]d","expiration":"%[5]s"}],"memo":"AAAAAB"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"create_offer"}]},{"type":"likechain.likenft.v1.EventCreateOffer","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"buyer","value":"\"%[1]s\""}]}]}]}`,
			ADDR_01_LIKE, nftClasses[0].Id, nfts[1].NftId, initPrice2, expiration.Add(1*time.Second).Format(time.RFC3339),
		),
	}
	blockTime := expiration.Add(-10000 * time.Second)
	InsertTestData(DBTestData{
		Iscns:           iscns,
		NftClasses:      nftClasses,
		Nfts:            nfts,
		Txs:             txs,
		LatestBlockTime: &blockTime,
	})

	ownersRes, err := GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 2)

	itemsRes, err := GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "offer"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, itemsRes.Items)

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 2)

	itemsRes, err = GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "offer"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, itemsRes.Items, 2)
	require.Equal(t, "offer", itemsRes.Items[0].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[0].ClassId)
	require.Equal(t, nfts[0].NftId, itemsRes.Items[0].NftId)
	require.Equal(t, ADDR_02_LIKE, itemsRes.Items[0].Creator)
	require.Equal(t, initPrice1, itemsRes.Items[0].Price)
	require.Equal(t, expiration, itemsRes.Items[0].Expiration)
	require.Equal(t, "offer", itemsRes.Items[1].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[1].ClassId)
	require.Equal(t, nfts[1].NftId, itemsRes.Items[1].NftId)
	require.Equal(t, ADDR_01_LIKE, itemsRes.Items[1].Creator)
	require.Equal(t, initPrice2, itemsRes.Items[1].Price)
	require.Equal(t, expiration.Add(1*time.Second), itemsRes.Items[1].Expiration)

	updatedPrice1 := uint64(100000000002)
	txs = []string{
		fmt.Sprintf(
			`{"txhash":"AAAAAC","height":"1236","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgUpdateOffer","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","price":"%[4]d","expiration":"%[5]s"}],"memo":"AAAAAC"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"update_offer"}]},{"type":"likechain.likenft.v1.EventUpdateOffer","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"buyer","value":"\"%[1]s\""}]}]}]}`,
			ADDR_02_LIKE, nftClasses[0].Id, nfts[0].NftId, updatedPrice1, expiration.Add(2*time.Second).Format(time.RFC3339),
		),
	}
	InsertTestData(DBTestData{Txs: txs})

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 2)

	itemsRes, err = GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "offer"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, itemsRes.Items, 2)
	require.Equal(t, "offer", itemsRes.Items[0].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[0].ClassId)
	require.Equal(t, nfts[1].NftId, itemsRes.Items[0].NftId)
	require.Equal(t, ADDR_01_LIKE, itemsRes.Items[0].Creator)
	require.Equal(t, initPrice2, itemsRes.Items[0].Price)
	require.Equal(t, expiration.Add(1*time.Second), itemsRes.Items[0].Expiration)
	require.Equal(t, "offer", itemsRes.Items[1].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[1].ClassId)
	require.Equal(t, nfts[0].NftId, itemsRes.Items[1].NftId)
	require.Equal(t, ADDR_02_LIKE, itemsRes.Items[1].Creator)
	require.Equal(t, updatedPrice1, itemsRes.Items[1].Price)
	require.Equal(t, expiration.Add(2*time.Second), itemsRes.Items[1].Expiration)

	txs = []string{
		fmt.Sprintf(
			`{"txhash":"AAAAAD","height":"1237","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgDeleteOffer","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s"}],"memo":"AAAAAD"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"delete_offer"}]},{"type":"likechain.likenft.v1.EventDeleteOffer","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"buyer","value":"\"%[1]s\""}]}]}]}`,
			ADDR_01_LIKE, nftClasses[0].Id, nfts[1].NftId,
		),
	}
	InsertTestData(DBTestData{Txs: txs})

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 2)

	itemsRes, err = GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "offer"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, itemsRes.Items, 1)
	require.Equal(t, "offer", itemsRes.Items[0].Type)
	require.Equal(t, nftClasses[0].Id, itemsRes.Items[0].ClassId)
	require.Equal(t, nfts[0].NftId, itemsRes.Items[0].NftId)
	require.Equal(t, ADDR_02_LIKE, itemsRes.Items[0].Creator)
	require.Equal(t, updatedPrice1, itemsRes.Items[0].Price)
	require.Equal(t, expiration.Add(2*time.Second), itemsRes.Items[0].Expiration)

	txs = []string{
		fmt.Sprintf(
			`{"txhash":"AAAAAE","height":"1238","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgSellNFT","creator":"%[1]s","class_id":"%[2]s","nft_id":"%[3]s","buyer":"%[4]s","price":"%[5]d"}],"memo":"AAAAAE"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"likechain.likenft.v1.EventSellNFT","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"buyer","value":"\"%[4]s\""},{"key":"seller","value":"\"%[1]s\""},{"key":"price","value":"\"%[5]d\""}]},{"type":"likechain.likenft.v1.EventDeleteOffer","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"buyer","value":"\"%[4]s\""}]},{"type":"message","attributes":[{"key":"action","value":"sell_nft"},{"key":"sender","value":"%[1]s"}]}]}]}`,
			ADDR_01_LIKE, nftClasses[0].Id, nfts[0].NftId, ADDR_02_LIKE, updatedPrice1,
		),
	}
	InsertTestData(DBTestData{Txs: txs})

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, ADDR_02_LIKE, ownersRes.Owners[0].Owner)
	require.Len(t, ownersRes.Owners[0].Nfts, 2)

	itemsRes, err = GetNftMarketplaceItems(Conn, QueryNftMarketplaceItemsRequest{Type: "offer"}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, itemsRes.Items)

	eventsRes, err := GetNftEvents(Conn,
		QueryEventsRequest{
			ClassId:    nftClasses[0].Id,
			NftId:      nfts[0].NftId,
			ActionType: []NftEventAction{ACTION_SELL},
		},
		PageRequest{Limit: 1, Reverse: true},
	)
	require.NoError(t, err)
	require.Len(t, eventsRes.Events, 1)
	require.Equal(t, updatedPrice1, eventsRes.Events[0].Price)
}
