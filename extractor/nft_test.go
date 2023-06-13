package extractor_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/extractor"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestCreateAndUpdateNft(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_01_LIKE,
		},
	}
	classId := "likenft1abcdef"
	name := "test nft class"
	symbol := "TEST"
	uri := "https://testing.com/aaaaaa"
	uriHash := "asdf"
	metadata := `{"a": "b", "c": "d"}`
	description := "testing NFT new class"
	timestamp := time.Unix(1234567890, 0).UTC()
	config := `{"burnable": true, "max_supply": "0", "blind_box_config": null}`
	txs := []string{
		fmt.Sprintf(`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"memo":"AAAAAA","messages":[{"@type":"/likechain.likenft.v1.MsgNewClass","input":{"name":"%[4]s","symbol":"%[5]s","uri":"%[6]s","uri_hash":"%[7]s","config":%[11]s,"metadata":%[8]s,"description":"%[9]s"},"parent":{"type":"ISCN","iscn_id_prefix":"%[2]s"},"creator":"%[1]s"}]}},"logs":[{"log":"","events":[{"type":"likechain.likenft.v1.EventNewClass","attributes":[{"key":"parent_iscn_id_prefix","value":"\"%[2]s\""},{"key":"parent_account","value":"\"\""},{"key":"class_id","value":"\"%[3]s\""}]},{"type":"message","attributes":[{"key":"action","value":"new_class"},{"key":"sender","value":"%[1]s"}]}],"msg_index":0}],"timestamp":"%[10]s"}`,
			ADDR_01_LIKE, prefixA, classId, name, symbol,
			uri, uriHash, metadata, description, timestamp.Format(time.RFC3339),
			config,
		),
	}
	InsertTestData(DBTestData{
		Iscns: iscns,
		Txs:   txs,
	})

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	pagination := PageRequest{Limit: 10}
	res, err := GetClasses(Conn, QueryClassRequest{}, pagination)
	require.NoError(t, err)
	require.Len(t, res.Classes, 1)

	require.Equal(t, classId, res.Classes[0].Id)
	require.Equal(t, name, res.Classes[0].Name)
	require.Equal(t, symbol, res.Classes[0].Symbol)
	require.Equal(t, uri, res.Classes[0].URI)
	require.Equal(t, uriHash, res.Classes[0].URIHash)
	require.Equal(t, description, res.Classes[0].Description)
	require.Equal(t, metadata, string(res.Classes[0].Metadata))
	require.Equal(t, timestamp, res.Classes[0].CreatedAt.UTC())
	require.Equal(t, config, string(res.Classes[0].Config))

	eventRes, err := GetNftEvents(Conn, QueryEventsRequest{
		ActionType: []NftEventAction{ACTION_NEW_CLASS},
		ClassId:    classId,
	}, pagination)
	require.NoError(t, err)
	require.Len(t, eventRes.Events, 1)
	require.Equal(t, ADDR_01_LIKE, eventRes.Events[0].Sender)
	require.Equal(t, "AAAAAA", eventRes.Events[0].Memo)

	name = "updated test nft class"
	symbol = "TEST-UPDATED"
	uri = "https://testing.com/updated"
	uriHash = "updated"
	metadata = `{"e": "f", "g": "h"}`
	description = "updated testing NFT new class"
	updateTimestamp := time.Unix(1234567891, 0).UTC()
	config = `{"burnable": false, "max_supply": "1", "blind_box_config": null}`
	txs = []string{
		fmt.Sprintf(`{"txhash":"AAAAAB","height":"1235","tx":{"body":{"memo":"AAAAAB","messages":[{"@type":"/likechain.likenft.v1.MsgUpdateClass","class_id":"%[3]s","input":{"name":"%[4]s","symbol":"%[5]s","uri":"%[6]s","uri_hash":"%[7]s","config":%[11]s,"metadata":%[8]s,"description":"%[9]s"},"creator":"%[1]s"}]}},"logs":[{"log":"","events":[{"type":"likechain.likenft.v1.EventUpdateClass","attributes":[{"key":"parent_iscn_id_prefix","value":"\"%[2]s\""},{"key":"parent_account","value":"\"\""},{"key":"class_id","value":"\"%[3]s\""}]},{"type":"message","attributes":[{"key":"action","value":"update_class"},{"key":"sender","value":"%[1]s"}]}],"msg_index":0}],"timestamp":"%[10]s"}`,
			ADDR_01_LIKE, prefixA, classId, name, symbol,
			uri, uriHash, metadata, description, updateTimestamp.Format(time.RFC3339),
			config,
		),
	}
	InsertTestData(DBTestData{
		Iscns: iscns,
		Txs:   txs,
	})

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	pagination = PageRequest{Limit: 10}
	res, err = GetClasses(Conn, QueryClassRequest{}, pagination)
	require.NoError(t, err)
	require.Len(t, res.Classes, 1)

	require.Equal(t, classId, res.Classes[0].Id)
	require.Equal(t, name, res.Classes[0].Name)
	require.Equal(t, symbol, res.Classes[0].Symbol)
	require.Equal(t, uri, res.Classes[0].URI)
	require.Equal(t, uriHash, res.Classes[0].URIHash)
	require.Equal(t, description, res.Classes[0].Description)
	require.Equal(t, metadata, string(res.Classes[0].Metadata))
	require.Equal(t, timestamp, res.Classes[0].CreatedAt.UTC())
	require.Equal(t, config, string(res.Classes[0].Config))

	eventRes, err = GetNftEvents(Conn, QueryEventsRequest{
		ActionType: []NftEventAction{ACTION_UPDATE_CLASS},
		ClassId:    classId,
	}, pagination)
	require.NoError(t, err)
	require.Len(t, eventRes.Events, 1)
	require.Equal(t, ADDR_01_LIKE, eventRes.Events[0].Sender)
	require.Equal(t, "AAAAAB", eventRes.Events[0].Memo)
}

func TestSendNft(t *testing.T) {
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
		fmt.Sprintf(`{"txhash":"AAAAAA","height":"1234","tx":{"body":{"messages":[{"@type":"/cosmos.nft.v1beta1.MsgSend","sender":"%[4]s","class_id":"%[2]s","id":"%[3]s","receiver":"%[1]s"}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"cosmos.nft.v1beta1.EventSend","attributes":[{"key":"class_id","value":"\"%[2]s\""},{"key":"id","value":"\"%[3]s\""},{"key":"sender","value":"\"%[4]s\""},{"key":"receiver","value":"\"%[1]s\""}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.nft.v1beta1.MsgSend"}]}]}]}`, ADDR_02_LIKE, nftClasses[0].Id, nfts[0].NftId, ADDR_01_LIKE),
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
	require.Equal(t, ACTION_SEND, eventRes.Events[0].Action)
	require.Equal(t, nfts[0].NftId, eventRes.Events[0].NftId)
	require.Equal(t, ADDR_01_LIKE, eventRes.Events[0].Sender)
	require.Equal(t, ADDR_02_LIKE, eventRes.Events[0].Receiver)
	require.Equal(t, "AAAAAA", eventRes.Events[0].TxHash)
	require.Equal(t, "AAAAAA", eventRes.Events[0].Memo)
}

func TestSendNftWithPrice(t *testing.T) {
	defer CleanupTestData(Conn)
	buyer := ADDR_02_LIKE
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
			Owner:   buyer,
		},
	}
	timestamp := time.Unix(1234567890, 0).UTC()
	iscnOwner := iscns[0].Owner
	apiWallet := ADDR_03_LIKE
	stakeholder := apiWallet
	price := uint64(100)
	royalty1 := uint64(78)
	royalty2 := price - royalty1
	txs := []string{
		fmt.Sprintf(`
{"height":"1234","txhash":"AAAAAA","logs":[{"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[2]s"},{"key":"amount","value":"%[8]dnanolike"},{"key":"authz_msg_index","value":"0"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[1]s"},{"key":"amount","value":"%[8]dnanolike"},{"key":"authz_msg_index","value":"0"}]},{"type":"cosmos.authz.v1beta1.EventRevoke","attributes":[{"key":"grantee","value":"\"%[2]s\""},{"key":"granter","value":"\"%[1]s\""},{"key":"msg_type_url","value":"\"/cosmos.bank.v1beta1.MsgSend\""}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.authz.v1beta1.MsgExec"},{"key":"sender","value":"%[1]s"},{"key":"authz_msg_index","value":"0"},{"key":"module","value":"bank"},{"key":"authz_msg_index","value":"0"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[2]s"},{"key":"sender","value":"%[1]s"},{"key":"amount","value":"%[8]dnanolike"},{"key":"authz_msg_index","value":"0"}]}]},{"events":[{"type":"cosmos.nft.v1beta1.EventSend","attributes":[{"key":"class_id","value":"\"%[5]s\""},{"key":"id","value":"\"%[6]s\""},{"key":"receiver","value":"\"%[1]s\""},{"key":"sender","value":"\"%[2]s\""}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.nft.v1beta1.MsgSend"}]}]},{"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[4]s"},{"key":"amount","value":"%[10]dnanolike"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[2]s"},{"key":"amount","value":"%[10]dnanolike"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.bank.v1beta1.MsgSend"},{"key":"sender","value":"%[2]s"},{"key":"module","value":"bank"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[4]s"},{"key":"sender","value":"%[2]s"},{"key":"amount","value":"%[10]dnanolike"}]}]},{"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[3]s"},{"key":"amount","value":"%[9]dnanolike"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[2]s"},{"key":"amount","value":"%[9]dnanolike"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.bank.v1beta1.MsgSend"},{"key":"sender","value":"%[2]s"},{"key":"module","value":"bank"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[3]s"},{"key":"sender","value":"%[2]s"},{"key":"amount","value":"%[9]dnanolike"}]}]}],"tx":{"body":{"messages":[{"@type":"/cosmos.authz.v1beta1.MsgExec","grantee":"%[2]s","msgs":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[1]s","to_address":"%[2]s","amount":[{"denom":"nanolike","amount":"%[8]d"}]}]},{"@type":"/cosmos.nft.v1beta1.MsgSend","class_id":"%[5]s","id":"%[6]s","sender":"%[2]s","receiver":"%[1]s"},{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[2]s","to_address":"%[3]s","amount":[{"denom":"nanolike","amount":"%[10]d"}]},{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[2]s","to_address":"%[4]s","amount":[{"denom":"nanolike","amount":"%[9]d"}]}],"memo":"AAAAAA"}},"timestamp":"%[7]s"}`,
			buyer, apiWallet, iscnOwner, stakeholder, nftClasses[0].Id,
			nfts[0].NftId, timestamp.Format(time.RFC3339), price, royalty1, royalty2),
	}
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		Txs:        txs,
	})

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err := GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, ownersRes.Owners[0].Owner, ADDR_02_LIKE)

	eventRes, err := GetNftEvents(Conn, QueryEventsRequest{
		ClassId: nftClasses[0].Id,
	}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, eventRes.Events, 1)
	require.Equal(t, nfts[0].NftId, eventRes.Events[0].NftId)
	require.Equal(t, apiWallet, eventRes.Events[0].Sender)
	require.Equal(t, buyer, eventRes.Events[0].Receiver)
	require.Equal(t, "AAAAAA", eventRes.Events[0].TxHash)
	require.Equal(t, ACTION_SEND, eventRes.Events[0].Action)
	require.Equal(t, price, eventRes.Events[0].Price)

	incomesRes, err := GetNftIncomes(Conn,
		QueryIncomesRequest{
			ClassId: nftClasses[0].Id,
		}, PageRequest{Limit: 10, Reverse: true},
	)
	require.NoError(t, err)
	classIncome := incomesRes.ClassIncomes[0]
	require.Equal(t, classIncome.ClassId, nftClasses[0].Id)
	require.Len(t, classIncome.Incomes, 2)
	require.Equal(t, iscnOwner, classIncome.Incomes[0].Address)
	require.Equal(t, royalty1, classIncome.Incomes[0].Amount)
	require.Equal(t, stakeholder, classIncome.Incomes[1].Address)
	require.Equal(t, royalty2, classIncome.Incomes[1].Amount)
	require.Equal(t, price, classIncome.TotalAmount)
	require.Equal(t, price, classIncome.Sales)
	require.Equal(t, incomesRes.TotalAmount, classIncome.TotalAmount)
	require.Equal(t, incomesRes.TotalSales, classIncome.Sales)

	row := Conn.QueryRow(context.Background(), `SELECT latest_price, price_updated_at FROM nft WHERE class_id = $1 AND nft_id = $2`, nftClasses[0].Id, nfts[0].NftId)
	var lastPrice uint64
	var priceUpdatedAt time.Time
	err = row.Scan(&lastPrice, &priceUpdatedAt)
	require.NoError(t, err)
	require.Equal(t, price, lastPrice)
	require.Equal(t, timestamp.UTC(), priceUpdatedAt.UTC())
}
func TestSendMultipleNftsWithPrice(t *testing.T) {
	defer CleanupTestData(Conn)
	buyer := ADDR_01_LIKE
	prefixA := "iscn://testing/aaaaaa"
	prefixB := "iscn://testing/bbbbbb"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_01_LIKE,
		},
		{
			Iscn:  "iscn://testing/aaaaaa/2",
			Owner: ADDR_02_LIKE,
		},
		{
			Iscn:  "iscn://testing/bbbbbb/1",
			Owner: ADDR_03_LIKE,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
		{
			Id:     "nftlike1bbbbb1",
			Parent: NftClassParent{IscnIdPrefix: prefixB},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-919775",
			ClassId: nftClasses[0].Id,
			Owner:   buyer,
		},
		{
			NftId:   "testing-nft-919776",
			ClassId: nftClasses[1].Id,
			Owner:   buyer,
		},
	}
	timestamp := time.Unix(1234567890, 0).UTC()
	apiWallet := ADDR_04_LIKE
	stakeholderA1 := ADDR_02_LIKE
	stakeholderA2 := apiWallet
	stakeholderB1 := ADDR_03_LIKE
	stakeholderB2 := apiWallet
	priceA := uint64(100)
	priceB := uint64(200)
	royaltyA1 := uint64(95)
	royaltyA2 := priceA - royaltyA1
	royaltyB1 := uint64(190)
	royaltyB2 := priceB - royaltyB1
	txs := []string{
		fmt.Sprintf(`
{"height":"1234","txhash":"AAAAAA","logs":[{"msg_index":0,"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[2]s"},{"key":"amount","value":"%[12]dnanolike"},{"key":"authz_msg_index","value":"0"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[1]s"},{"key":"amount","value":"%[12]dnanolike"},{"key":"authz_msg_index","value":"0"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.authz.v1beta1.MsgExec"},{"key":"sender","value":"%[1]s"},{"key":"authz_msg_index","value":"0"},{"key":"module","value":"bank"},{"key":"authz_msg_index","value":"0"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[1]s"},{"key":"sender","value":"%[2]s"},{"key":"amount","value":"%[12]dnanolike"},{"key":"authz_msg_index","value":"0"}]}]},{"msg_index":1,"events":[{"type":"cosmos.nft.v1beta1.EventSend","attributes":[{"key":"class_id","value":"\"%[7]s\""},{"key":"id","value":"\"%[9]s\""},{"key":"receiver","value":"\"%[1]s\""},{"key":"sender","value":"\"%[2]s\""}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.nft.v1beta1.MsgSend"}]}]},{"msg_index":2,"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[4]s"},{"key":"amount","value":"%[15]dnanolike"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[2]s"},{"key":"amount","value":"%[15]dnanolike"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.bank.v1beta1.MsgSend"},{"key":"sender","value":"%[2]s"},{"key":"module","value":"bank"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[4]s"},{"key":"sender","value":"%[2]s"},{"key":"amount","value":"%[15]dnanolike"}]}]},{"msg_index":3,"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[3]s"},{"key":"amount","value":"%[14]dnanolike"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[2]s"},{"key":"amount","value":"%[14]dnanolike"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.bank.v1beta1.MsgSend"},{"key":"sender","value":"%[2]s"},{"key":"module","value":"bank"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[3]s"},{"key":"sender","value":"%[2]s"},{"key":"amount","value":"%[14]dnanolike"}]}]},{"msg_index":4,"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[2]s"},{"key":"amount","value":"%[13]dnanolike"},{"key":"authz_msg_index","value":"0"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[1]s"},{"key":"amount","value":"%[13]dnanolike"},{"key":"authz_msg_index","value":"0"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.authz.v1beta1.MsgExec"},{"key":"sender","value":"%[1]s"},{"key":"authz_msg_index","value":"0"},{"key":"module","value":"bank"},{"key":"authz_msg_index","value":"0"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[2]s"},{"key":"sender","value":"%[1]s"},{"key":"amount","value":"%[13]dnanolike"},{"key":"authz_msg_index","value":"0"}]}]},{"msg_index":5,"events":[{"type":"cosmos.nft.v1beta1.EventSend","attributes":[{"key":"class_id","value":"\"%[8]s\""},{"key":"id","value":"\"%[10]s\""},{"key":"receiver","value":"\"%[1]s\""},{"key":"sender","value":"\"%[2]s\""}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.nft.v1beta1.MsgSend"}]}]},{"msg_index":6,"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[6]s"},{"key":"amount","value":"%[17]dnanolike"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[2]s"},{"key":"amount","value":"%[17]dnanolike"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.bank.v1beta1.MsgSend"},{"key":"sender","value":"%[2]s"},{"key":"module","value":"bank"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[6]s"},{"key":"sender","value":"%[2]s"},{"key":"amount","value":"%[17]dnanolike"}]}]},{"msg_index":7,"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"%[5]s"},{"key":"amount","value":"%[16]dnanolike"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"%[2]s"},{"key":"amount","value":"%[16]dnanolike"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.bank.v1beta1.MsgSend"},{"key":"sender","value":"%[2]s"},{"key":"module","value":"bank"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"%[5]s"},{"key":"sender","value":"%[2]s"},{"key":"amount","value":"%[16]dnanolike"}]}]}],"tx":{"body":{"messages":[{"@type":"/cosmos.authz.v1beta1.MsgExec","grantee":"%[2]s","msgs":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[1]s","to_address":"%[2]s","amount":[{"denom":"nanolike","amount":"%[12]d"}]}]},{"@type":"/cosmos.nft.v1beta1.MsgSend","class_id":"%[7]s","id":"%[9]s","sender":"%[2]s","receiver":"%[1]s"},{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[2]s","to_address":"%[4]s","amount":[{"denom":"nanolike","amount":"%[15]d"}]},{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[2]s","to_address":"%[3]s","amount":[{"denom":"nanolike","amount":"%[14]d"}]},{"@type":"/cosmos.authz.v1beta1.MsgExec","grantee":"%[2]s","msgs":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[1]s","to_address":"%[2]s","amount":[{"denom":"nanolike","amount":"%[13]d"}]}]},{"@type":"/cosmos.nft.v1beta1.MsgSend","class_id":"%[8]s","id":"%[10]s","sender":"%[2]s","receiver":"%[1]s"},{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[2]s","to_address":"%[6]s","amount":[{"denom":"nanolike","amount":"%[17]d"}]},{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%[2]s","to_address":"%[5]s","amount":[{"denom":"nanolike","amount":"%[16]d"}]}],"memo":"(multiple purchases)"}},"timestamp":"%[11]s"}`,
			buyer, apiWallet, stakeholderA1, stakeholderA2, stakeholderB1,
			stakeholderB2, nftClasses[0].Id, nftClasses[1].Id, nfts[0].NftId, nfts[1].NftId,
			timestamp.Format(time.RFC3339), priceA, priceB, royaltyA1, royaltyA2,
			royaltyB1, royaltyB2,
		),
	}
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		Txs:        txs,
	})

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	ownersRes, err := GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[0].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, buyer, ownersRes.Owners[0].Owner)
	ownersRes, err = GetOwners(Conn, QueryOwnerRequest{
		ClassId: nftClasses[1].Id,
	})
	require.NoError(t, err)
	require.Len(t, ownersRes.Owners, 1)
	require.Equal(t, buyer, ownersRes.Owners[0].Owner)

	eventRes, err := GetNftEvents(Conn, QueryEventsRequest{
		ClassId: nftClasses[0].Id,
	}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, eventRes.Events, 1)
	nftEvent := eventRes.Events[0]
	require.Equal(t, nfts[0].NftId, nftEvent.NftId)
	require.Equal(t, apiWallet, nftEvent.Sender)
	require.Equal(t, buyer, nftEvent.Receiver)
	require.Equal(t, "AAAAAA", nftEvent.TxHash)
	require.Equal(t, ACTION_SEND, nftEvent.Action)
	require.Equal(t, priceA, nftEvent.Price)
	eventRes, err = GetNftEvents(Conn, QueryEventsRequest{
		ClassId: nftClasses[1].Id,
	}, PageRequest{Limit: 10})
	require.NoError(t, err)
	require.Len(t, eventRes.Events, 1)
	nftEvent = eventRes.Events[0]
	require.Equal(t, nfts[1].NftId, nftEvent.NftId)
	require.Equal(t, apiWallet, nftEvent.Sender)
	require.Equal(t, buyer, nftEvent.Receiver)
	require.Equal(t, "AAAAAA", nftEvent.TxHash)
	require.Equal(t, ACTION_SEND, nftEvent.Action)
	require.Equal(t, priceB, nftEvent.Price)

	incomesRes, err := GetNftIncomes(Conn,
		QueryIncomesRequest{
			ClassId: nftClasses[0].Id,
		}, PageRequest{Limit: 10, Reverse: true},
	)
	require.NoError(t, err)
	classIncome := incomesRes.ClassIncomes[0]
	require.Equal(t, classIncome.ClassId, nftClasses[0].Id)
	require.Len(t, classIncome.Incomes, 2)
	require.Equal(t, stakeholderA1, classIncome.Incomes[0].Address)
	require.Equal(t, royaltyA1, classIncome.Incomes[0].Amount)
	require.Equal(t, stakeholderA2, classIncome.Incomes[1].Address)
	require.Equal(t, royaltyA2, classIncome.Incomes[1].Amount)
	require.Equal(t, priceA, classIncome.TotalAmount)
	require.Equal(t, priceA, classIncome.Sales)
	require.Equal(t, incomesRes.TotalAmount, classIncome.TotalAmount)
	require.Equal(t, incomesRes.TotalSales, classIncome.Sales)
	incomesRes, err = GetNftIncomes(Conn,
		QueryIncomesRequest{
			ClassId: nftClasses[1].Id,
		}, PageRequest{Limit: 10, Reverse: true},
	)
	require.NoError(t, err)
	classIncome = incomesRes.ClassIncomes[0]
	require.Equal(t, classIncome.ClassId, nftClasses[1].Id)
	require.Len(t, classIncome.Incomes, 2)
	require.Equal(t, stakeholderB1, classIncome.Incomes[0].Address)
	require.Equal(t, royaltyB1, classIncome.Incomes[0].Amount)
	require.Equal(t, stakeholderB2, classIncome.Incomes[1].Address)
	require.Equal(t, royaltyB2, classIncome.Incomes[1].Amount)
	require.Equal(t, priceB, classIncome.TotalAmount)
	require.Equal(t, priceB, classIncome.Sales)
	require.Equal(t, incomesRes.TotalAmount, classIncome.TotalAmount)
	require.Equal(t, incomesRes.TotalSales, classIncome.Sales)

	row := Conn.QueryRow(context.Background(), `SELECT latest_price, price_updated_at FROM nft WHERE class_id = $1 AND nft_id = $2`, nftClasses[0].Id, nfts[0].NftId)
	var lastPrice uint64
	var priceUpdatedAt time.Time
	err = row.Scan(&lastPrice, &priceUpdatedAt)
	require.NoError(t, err)
	require.Equal(t, priceA, lastPrice)
	require.Equal(t, timestamp.UTC(), priceUpdatedAt.UTC())
	row = Conn.QueryRow(context.Background(), `SELECT latest_price, price_updated_at FROM nft WHERE class_id = $1 AND nft_id = $2`, nftClasses[1].Id, nfts[1].NftId)
	err = row.Scan(&lastPrice, &priceUpdatedAt)
	require.NoError(t, err)
	require.Equal(t, priceB, lastPrice)
	require.Equal(t, timestamp.UTC(), priceUpdatedAt.UTC())
}

func TestMintNft(t *testing.T) {
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
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Txs:        txs,
	})

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	// hack: since currently GetNfts requires event with receiver = owner,
	// we insert a dummy event here for testing purpose
	InsertTestData(DBTestData{
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
			ActionType: []NftEventAction{ACTION_MINT},
		},
		PageRequest{Limit: 1, Reverse: true},
	)
	require.NoError(t, err)
	require.Len(t, eventsRes.Events, 1)
	require.Equal(t, ADDR_01_LIKE, eventsRes.Events[0].Sender)
	require.Equal(t, "AAAAAA", eventsRes.Events[0].Memo)
}

func TestNftEventIscnOwner(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_02_LIKE,
		},
		{
			Iscn:  "iscn://testing/aaaaaa/2",
			Owner: ADDR_01_LIKE,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
		{
			Id:     "nftlike1aaaaa2",
			Parent: NftClassParent{IscnIdPrefix: "no_such_iscn"},
		},
	}
	noSuchClass := "no_such_class"

	nftId := "testing-nft-1092934"
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
		fmt.Sprintf(`
		{"txhash":"AAAAAB","height":"1234","tx":{"body":{"memo":"AAAAAB","messages":[{"id":"%[1]s","@type":"/likechain.likenft.v1.MsgMintNFT","input":{"uri":"%[2]s","uri_hash":"%[3]s","metadata":%[4]s},"creator":"%[5]s","class_id":"%[6]s"}]}},"logs":[{"events":[{"type":"cosmos.nft.v1beta1.EventMint","attributes":[{"key":"id","value":"\"%[1]s\""},{"key":"owner","value":"\"%[5]s\""},{"key":"class_id","value":"\"%[6]s\""}]},{"type":"likechain.likenft.v1.EventMintNFT","attributes":[{"key":"class_id","value":"\"%[6]s\""},{"key":"nft_id","value":"\"%[1]s\""},{"key":"owner","value":"\"%[5]s\""},{"key":"class_parent_iscn_id_prefix","value":"\"%[7]s\""},{"key":"class_parent_account","value":"\"\""}]},{"type":"message","attributes":[{"key":"action","value":"mint_nft"},{"key":"sender","value":"%[5]s"}]}],"msg_index":0}],"timestamp":"%[8]s"}`,
			nftId, uri, uriHash, metadata, ADDR_01_LIKE,
			nftClasses[1].Id, prefixA, timestamp.Format(time.RFC3339),
		),
		fmt.Sprintf(`
		{"txhash":"AAAAAC","height":"1234","tx":{"body":{"memo":"AAAAAC","messages":[{"id":"%[1]s","@type":"/likechain.likenft.v1.MsgMintNFT","input":{"uri":"%[2]s","uri_hash":"%[3]s","metadata":%[4]s},"creator":"%[5]s","class_id":"%[6]s"}]}},"logs":[{"events":[{"type":"cosmos.nft.v1beta1.EventMint","attributes":[{"key":"id","value":"\"%[1]s\""},{"key":"owner","value":"\"%[5]s\""},{"key":"class_id","value":"\"%[6]s\""}]},{"type":"likechain.likenft.v1.EventMintNFT","attributes":[{"key":"class_id","value":"\"%[6]s\""},{"key":"nft_id","value":"\"%[1]s\""},{"key":"owner","value":"\"%[5]s\""},{"key":"class_parent_iscn_id_prefix","value":"\"%[7]s\""},{"key":"class_parent_account","value":"\"\""}]},{"type":"message","attributes":[{"key":"action","value":"mint_nft"},{"key":"sender","value":"%[5]s"}]}],"msg_index":0}],"timestamp":"%[8]s"}`,
			nftId, uri, uriHash, metadata, ADDR_01_LIKE,
			noSuchClass, prefixA, timestamp.Format(time.RFC3339),
		),
	}
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Txs:        txs,
	})

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	var iscnOwner string

	row := Conn.QueryRow(
		context.Background(),
		`SELECT iscn_owner_at_the_time FROM nft_event WHERE nft_id = $1 AND class_id = $2`,
		nftId, nftClasses[0].Id,
	)
	err = row.Scan(&iscnOwner)
	require.NoError(t, err)
	require.Equal(t, ADDR_01_LIKE, iscnOwner)

	row = Conn.QueryRow(
		context.Background(),
		`SELECT iscn_owner_at_the_time FROM nft_event WHERE nft_id = $1 AND class_id = $2`,
		nftId, nftClasses[1].Id,
	)
	err = row.Scan(&iscnOwner)
	require.NoError(t, err)
	require.Equal(t, "", iscnOwner)

	row = Conn.QueryRow(
		context.Background(),
		`SELECT iscn_owner_at_the_time FROM nft_event WHERE nft_id = $1 AND class_id = $2`,
		nftId, noSuchClass,
	)
	err = row.Scan(&iscnOwner)
	require.NoError(t, err)
	require.Equal(t, "", iscnOwner)
}
