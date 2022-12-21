package extractor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestBuyNft(t *testing.T) {
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
	err := PrepareTestData(iscns, nftClasses, nfts, nil, txs)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

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

	finished, err := Extract(Conn, handlers)
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
	require.Equal(t, eventRes.Events[0].Action, "buy_nft")
	require.Equal(t, eventRes.Events[0].NftId, nfts[0].NftId)
	require.Equal(t, eventRes.Events[0].Sender, ADDR_01_LIKE)
	require.Equal(t, eventRes.Events[0].Receiver, ADDR_02_LIKE)
	require.Equal(t, eventRes.Events[0].TxHash, "AAAAAA")
}

func TestSellNft(t *testing.T) {
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
		`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"memo":"AAAAAA"}}}`,
		fmt.Sprintf(`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"messages":[{"@type":"/likechain.likenft.v1.MsgSellNFT","creator":"%[4]s","class_id":"%[2]s","nft_id":"%[3]s","buyer":"%[1]s","price":"100000000000"}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"likechain.likenft.v1.EventSellNFT","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"nft_id","value":"\"%[3]s\""},{"key":"seller","value":"\"%[4]s\""},{"key":"buyer","value":"\"%[1]s\""},{"key":"price","value":"\"100000000000\""}]},{"type":"message","attributes":[{"key":"action","value":"sell_nft"},{"key":"sender","value":"%[4]s"}]}]}]}`, ADDR_02_LIKE, nftClasses[0].Id, nfts[0].NftId, ADDR_01_LIKE),
	}
	err := PrepareTestData(iscns, nftClasses, nfts, nil, txs)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

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

	finished, err := Extract(Conn, handlers)
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
	require.Equal(t, eventRes.Events[0].Action, "sell_nft")
	require.Equal(t, eventRes.Events[0].NftId, nfts[0].NftId)
	require.Equal(t, eventRes.Events[0].Sender, ADDR_01_LIKE)
	require.Equal(t, eventRes.Events[0].Receiver, ADDR_02_LIKE)
	require.Equal(t, eventRes.Events[0].TxHash, "AAAAAA")
}
