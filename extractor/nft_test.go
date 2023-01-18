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

func TestSendNft(t *testing.T) {
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
		fmt.Sprintf(`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"messages":[{"@type":"/cosmos.nft.v1beta1.MsgSend","sender":"%[4]s","class_id":"%[2]s","id":"%[3]s","receiver":"%[1]s"}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"cosmos.nft.v1beta1.EventSend","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"id","value":"\"%[3]s\""},{"key":"sender","value":"\"%[4]s\""},{"key":"receiver","value":"\"%[1]s\""}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.nft.v1beta1.MsgSend"}]}]}]}`, ADDR_02_LIKE, nftClasses[0].Id, nfts[0].NftId, ADDR_01_LIKE),
	}
	err := InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		Txs:        txs,
	})
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
	require.Equal(t, eventRes.Events[0].Action, "/cosmos.nft.v1beta1.MsgSend")
	require.Equal(t, eventRes.Events[0].NftId, nfts[0].NftId)
	require.Equal(t, eventRes.Events[0].Sender, ADDR_01_LIKE)
	require.Equal(t, eventRes.Events[0].Receiver, ADDR_02_LIKE)
	require.Equal(t, eventRes.Events[0].TxHash, "AAAAAA")
}

func TestMintNft(t *testing.T) {
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

	nftId := "testing-nft-199920"
	uri := "https://testing.com/aaaaaa"
	uriHash := "asdf"
	metadata := `{"a": "b", "c": "d"}`
	timestamp := time.Unix(1234567890, 0).UTC()
	txs := []string{
		fmt.Sprintf(`
		{"txhash":"AAAAAA","height":"1234","tx":{"body":{"memo":"AAAAAA","messages":[{"id":"%[1]s","@type":"/likechain.likenft.v1.MsgMintNFT","input":{"uri":"%[2]s","uri_hash":"%[3]s","metadata":%[4]s},"creator":"%[5]s","class_id":"%[6]s"}]}},"logs":[{"events":[{"type":"cosmos.nft.v1beta1.EventMint","attributes":[{"key":"id","value":"\"%[1]s\""},{"key":"owner","value":"\"%[5]s\""},{"key":"class_id","value":"\"%[6]s\""}]},{"type":"likechain.likenft.v1.EventMintNFT","attributes":[{"key":"class_id","value":"\"%[6]s\""},{"key":"nft_id","value":"\"%[1]s\""},{"key":"owner","value":"\"%[5]s\""},{"key":"class_parent_iscn_id_prefix","value":"\"%[7]s\""},{"key":"class_parent_account","value":"\"\""}]},{"type":"message","attributes":[{"key":"action","value":"mint_nft"},{"key":"sender","value":"%[5]s"}]}],"msg_index":0}],"timestamp":"%[8]s"}`,
			nftId, uri, uriHash, metadata, ADDR_01_LIKE,
			nftClasses[0].Id, prefixA, timestamp.Format(time.RFC3339),
		),
	}
	err := InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Txs:        txs,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	// hack: since currently GetNfts requires event with receiver = owner,
	// we insert a dummy event here for testing purpose
	err = InsertTestData(DBTestData{
		NftEvents: []NftEvent{
			{
				Action:    "dummy",
				ClassId:   nftClasses[0].Id,
				NftId:     nftId,
				Sender:    ADDR_01_LIKE,
				Receiver:  ADDR_01_LIKE,
				Timestamp: timestamp,
			},
		},
	})
	require.NoError(t, err)

	res, err := GetNfts(Conn, QueryNftRequest{Owner: ADDR_01_LIKE}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, res.Nfts, 1)

	require.Equal(t, nftId, res.Nfts[0].NftId)
	require.Equal(t, uri, res.Nfts[0].Uri)
	require.Equal(t, uriHash, res.Nfts[0].UriHash)
	require.Equal(t, metadata, string(res.Nfts[0].Metadata))
	require.Equal(t, timestamp, res.Nfts[0].Timestamp)
	require.Equal(t, nftClasses[0].Id, res.Nfts[0].ClassId)

	eventsRes, err := GetNftEvents(Conn,
		QueryEventsRequest{
			ClassId:    nftClasses[0].Id,
			NftId:      nftId,
			ActionType: []string{"mint_nft"},
		},
		PageRequest{Limit: 1, Reverse: true},
	)
	require.NoError(t, err)
	require.Len(t, eventsRes.Events, 1)
	require.Equal(t, ADDR_01_LIKE, eventsRes.Events[0].Sender)
}
