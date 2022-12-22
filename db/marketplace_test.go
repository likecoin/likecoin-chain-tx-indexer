package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestMarketplace(t *testing.T) {
	prefixA := "iscn://testing/aaaaaa"
	iscns := []IscnInsert{
		{
			Iscn:       "iscn://testing/aaaaaa/1",
			IscnPrefix: prefixA,
			Owner:      ADDR_01_LIKE,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
		{
			Id:     "nftlike1bbbbbb",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-91301",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-91302",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_02_LIKE,
		},
		{
			NftId:   "testing-nft-91303",
			ClassId: nftClasses[1].Id,
			Owner:   ADDR_03_LIKE,
		},
		{
			NftId:   "testing-nft-91304",
			ClassId: nftClasses[1].Id,
			Owner:   ADDR_04_LIKE,
		},
	}
	expiration := time.Unix(1700000000, 0).UTC()
	marketplaceItems := []NftMarketplaceItem{
		{
			Type:       "listing",
			ClassId:    nftClasses[0].Id,
			NftId:      nfts[0].NftId,
			Creator:    ADDR_01_LIKE,
			Price:      100000000000,
			Expiration: expiration,
		},
		{
			Type:       "listing",
			ClassId:    nftClasses[0].Id,
			NftId:      nfts[1].NftId,
			Creator:    ADDR_02_LIKE,
			Price:      100000000001,
			Expiration: expiration.Add(1 * time.Second),
		},
		{
			Type:       "listing",
			ClassId:    nftClasses[1].Id,
			NftId:      nfts[2].NftId,
			Creator:    ADDR_03_LIKE,
			Price:      100000000002,
			Expiration: expiration.Add(2 * time.Second),
		},
		{
			Type:       "listing",
			ClassId:    nftClasses[1].Id,
			NftId:      nfts[3].NftId,
			Creator:    ADDR_04_LIKE,
			Price:      100000000003,
			Expiration: expiration.Add(-10000 * time.Second),
		},
		{
			Type:       "offer",
			ClassId:    nftClasses[0].Id,
			NftId:      nfts[1].NftId,
			Creator:    ADDR_01_LIKE,
			Price:      100000000004,
			Expiration: expiration,
		},
		{
			Type:       "offer",
			ClassId:    nftClasses[1].Id,
			NftId:      nfts[2].NftId,
			Creator:    ADDR_02_LIKE,
			Price:      100000000005,
			Expiration: expiration.Add(-10000 * time.Second),
		},
		{
			Type:       "offer",
			ClassId:    nftClasses[1].Id,
			NftId:      nfts[3].NftId,
			Creator:    ADDR_03_LIKE,
			Price:      100000000006,
			Expiration: expiration.Add(1 * time.Second),
		},
		{
			Type:       "offer",
			ClassId:    nftClasses[0].Id,
			NftId:      nfts[0].NftId,
			Creator:    ADDR_04_LIKE,
			Price:      100000000007,
			Expiration: expiration.Add(2 * time.Second),
		},
	}
	blockTime := expiration.Add(-100 * time.Second)
	err := InsertTestData(DBTestData{
		Iscns:               iscns,
		NftClasses:          nftClasses,
		Nfts:                nfts,
		NftMarketplaceItems: marketplaceItems,
		LatestBlockTime:     &blockTime,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

	table := []struct {
		name       string
		query      QueryNftMarketplaceItemsRequest
		page       PageRequest
		shouldFail bool
		length     int
		response   QueryNftMarketplaceItemsResponse
	}{
		{
			name:   "query all listings",
			query:  QueryNftMarketplaceItemsRequest{Type: "listing"},
			length: 3,
			response: QueryNftMarketplaceItemsResponse{
				Items: []NftMarketplaceItem{marketplaceItems[0], marketplaceItems[1], marketplaceItems[2]},
			},
		},
		{
			name:   "query all offers",
			query:  QueryNftMarketplaceItemsRequest{Type: "offer"},
			length: 3,
			response: QueryNftMarketplaceItemsResponse{
				Items: []NftMarketplaceItem{marketplaceItems[4], marketplaceItems[6], marketplaceItems[7]},
			},
		},
		{
			name:     "listings by creator",
			query:    QueryNftMarketplaceItemsRequest{Type: "listing", Creator: ADDR_01_LIKE},
			length:   1,
			response: QueryNftMarketplaceItemsResponse{Items: []NftMarketplaceItem{marketplaceItems[0]}},
		},
		{
			name:     "offers by creator",
			query:    QueryNftMarketplaceItemsRequest{Type: "offer", Creator: ADDR_01_LIKE},
			length:   1,
			response: QueryNftMarketplaceItemsResponse{Items: []NftMarketplaceItem{marketplaceItems[4]}},
		},
		{
			name:   "listings by class ID",
			query:  QueryNftMarketplaceItemsRequest{Type: "listing", ClassId: nftClasses[0].Id},
			length: 2,
			response: QueryNftMarketplaceItemsResponse{
				Items: []NftMarketplaceItem{marketplaceItems[0], marketplaceItems[1]},
			},
		},
		{
			name:     "offers by class ID",
			query:    QueryNftMarketplaceItemsRequest{Type: "offer", ClassId: nftClasses[1].Id},
			length:   1,
			response: QueryNftMarketplaceItemsResponse{Items: []NftMarketplaceItem{marketplaceItems[6]}},
		},
		{
			name:     "listings by NFT ID",
			query:    QueryNftMarketplaceItemsRequest{Type: "listing", NftId: nfts[1].NftId},
			length:   1,
			response: QueryNftMarketplaceItemsResponse{Items: []NftMarketplaceItem{marketplaceItems[1]}},
		},
		{
			name:     "offers by NFT ID",
			query:    QueryNftMarketplaceItemsRequest{Type: "offer", NftId: nfts[1].NftId},
			length:   1,
			response: QueryNftMarketplaceItemsResponse{Items: []NftMarketplaceItem{marketplaceItems[4]}},
		},
		{
			name:   "limit",
			query:  QueryNftMarketplaceItemsRequest{Type: "listing"},
			page:   PageRequest{Limit: 2},
			length: 2,
			response: QueryNftMarketplaceItemsResponse{
				Items: []NftMarketplaceItem{marketplaceItems[0], marketplaceItems[1]},
				Pagination: PageResponse{
					NextKey: uint64(marketplaceItems[1].Expiration.UnixNano()),
				},
			},
		},
		{
			name:     "after",
			query:    QueryNftMarketplaceItemsRequest{Type: "listing"},
			page:     PageRequest{Limit: 10, Key: uint64(marketplaceItems[1].Expiration.UnixNano())},
			length:   1,
			response: QueryNftMarketplaceItemsResponse{Items: []NftMarketplaceItem{marketplaceItems[2]}},
		},
		{
			name:   "reverse",
			query:  QueryNftMarketplaceItemsRequest{Type: "listing"},
			page:   PageRequest{Limit: 10, Reverse: true},
			length: 3,
			response: QueryNftMarketplaceItemsResponse{
				Items: []NftMarketplaceItem{marketplaceItems[2], marketplaceItems[1], marketplaceItems[0]},
			},
		},
	}
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			if v.page.Limit == 0 {
				v.page.Limit = 10
			}
			res, err := GetNftMarketplaceItems(Conn, v.query, v.page)
			require.NoError(t, err)
			require.Len(t, res.Items, v.length)
			for i, item := range v.response.Items {
				require.Equal(t, item, res.Items[i])
			}
			if v.response.Pagination.NextKey != 0 {
				require.Equal(t, v.response.Pagination.NextKey, res.Pagination.NextKey)
			}
		})
	}
}
