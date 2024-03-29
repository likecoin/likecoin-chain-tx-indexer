package rest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/rest"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestMarketplace(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	iscns := []db.IscnInsert{
		{
			Iscn:       "iscn://testing/aaaaaa/1",
			IscnPrefix: prefixA,
			Owner:      ADDR_01_LIKE,
		},
	}
	nftClasses := []db.NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: db.NftClassParent{IscnIdPrefix: prefixA},
		},
		{
			Id:       "nftlike1bbbbbb",
			Parent:   db.NftClassParent{IscnIdPrefix: prefixA},
			Metadata: json.RawMessage(`"name=a"`),
		},
	}
	nfts := []db.Nft{
		{
			NftId:    "testing-nft-91301",
			ClassId:  nftClasses[0].Id,
			Owner:    ADDR_01_LIKE,
			Metadata: json.RawMessage(`"name=91301"`),
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
	marketplaceItems := []db.NftMarketplaceItem{
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
	InsertTestData(DBTestData{
		Iscns:               iscns,
		NftClasses:          nftClasses,
		Nfts:                nfts,
		NftMarketplaceItems: marketplaceItems,
		LatestBlockTime:     &blockTime,
	})

	table := []struct {
		name          string
		query         string
		shouldFail    bool
		length        int
		items         []db.NftMarketplaceItem
		pagination    db.PageResponse
		classMetadata []json.RawMessage
		nftMetadata   []json.RawMessage
	}{
		{
			name:   "query all listings",
			query:  "type=listing",
			length: 3,
			items:  []db.NftMarketplaceItem{marketplaceItems[0], marketplaceItems[1], marketplaceItems[2]},
		},
		{
			name:   "query all offers",
			query:  "type=offer",
			length: 3,
			items:  []db.NftMarketplaceItem{marketplaceItems[4], marketplaceItems[6], marketplaceItems[7]},
		},
		{
			name:       "no type",
			query:      "",
			shouldFail: true,
		},
		{
			name:       "invalid type",
			query:      "type=listings",
			shouldFail: true,
		},
		{
			name:   "listings by creator",
			query:  "type=listing&creator=" + ADDR_01_LIKE,
			length: 1,
			items:  []db.NftMarketplaceItem{marketplaceItems[0]},
		},
		{
			name:   "offers by creator",
			query:  "type=offer&creator=" + ADDR_01_LIKE,
			length: 1,
			items:  []db.NftMarketplaceItem{marketplaceItems[4]},
		},
		{
			name:   "listings by class ID",
			query:  "type=listing&class_id=" + nftClasses[0].Id,
			length: 2,
			items:  []db.NftMarketplaceItem{marketplaceItems[0], marketplaceItems[1]},
		},
		{
			name:   "offers by class ID",
			query:  "type=offer&class_id=" + nftClasses[1].Id,
			length: 1,
			items:  []db.NftMarketplaceItem{marketplaceItems[6]},
		},
		{
			name:   "listings by NFT ID",
			query:  "type=listing&class_id=" + nftClasses[0].Id + "&nft_id=" + nfts[1].NftId,
			length: 1,
			items:  []db.NftMarketplaceItem{marketplaceItems[1]},
		},
		{
			name:   "offers by NFT ID",
			query:  "type=offer&class_id=" + nftClasses[0].Id + "&nft_id=" + nfts[1].NftId,
			length: 1,
			items:  []db.NftMarketplaceItem{marketplaceItems[4]},
		},
		{
			name:   "limit",
			query:  "type=listing&pagination.limit=2",
			length: 2,
			items:  []db.NftMarketplaceItem{marketplaceItems[0], marketplaceItems[1]},
			pagination: db.PageResponse{
				NextKey: uint64(marketplaceItems[1].Expiration.UnixNano()),
			},
		},
		{
			name:   "after",
			query:  fmt.Sprintf("type=listing&pagination.key=%d", marketplaceItems[1].Expiration.UnixNano()),
			length: 1,
			items:  []db.NftMarketplaceItem{marketplaceItems[2]},
		},
		{
			name:   "reverse",
			query:  "type=listing&pagination.reverse=true",
			length: 3,
			items:  []db.NftMarketplaceItem{marketplaceItems[2], marketplaceItems[1], marketplaceItems[0]},
		},
		{
			name:          "expand",
			query:         "type=listing&expand=true",
			length:        3,
			items:         []db.NftMarketplaceItem{marketplaceItems[0], marketplaceItems[1], marketplaceItems[2]},
			classMetadata: []json.RawMessage{nftClasses[0].Metadata, nftClasses[0].Metadata, nftClasses[1].Metadata},
			nftMetadata:   []json.RawMessage{nfts[0].Metadata, nfts[1].Metadata, nfts[2].Metadata},
		},
	}
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s/marketplace?%s", NFT_ENDPOINT, v.query), nil)
			httpRes, body := request(req)
			failed := (httpRes.StatusCode != 200)
			require.Equal(t, v.shouldFail, failed, "%##v", httpRes)
			if failed {
				return
			}
			var res db.QueryNftMarketplaceItemsResponse
			err := json.Unmarshal([]byte(body), &res)
			require.NoError(t, err)
			require.Len(t, res.Items, v.length)
			for i, item := range v.items {
				require.Equal(t, item.ClassId, res.Items[i].ClassId)
				require.Equal(t, item.NftId, res.Items[i].NftId)
				require.Equal(t, item.Creator, res.Items[i].Creator)
				require.Equal(t, item.Price, res.Items[i].Price)
				require.Equal(t, item.Expiration, res.Items[i].Expiration)
			}
			for i, classMetadata := range v.classMetadata {
				if classMetadata != nil {
					require.Equal(t, 0, bytes.Compare(classMetadata, res.Items[i].ClassMetadata), "%s <-> %s", classMetadata, res.Items[i].ClassMetadata)
				}
			}
			for i, nftMetadata := range v.nftMetadata {
				if nftMetadata != nil {
					require.Equal(t, 0, bytes.Compare(nftMetadata, res.Items[i].NftMetadata), "%s <-> %s", nftMetadata, res.Items[i].NftMetadata)
				}
			}
			if v.pagination.NextKey != 0 {
				require.Equal(t, v.pagination.NextKey, res.Pagination.NextKey)
			}
		})
	}
}
