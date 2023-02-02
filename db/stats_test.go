package db_test

import (
	"testing"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestNftCount(t *testing.T) {
	defer CleanupTestData(Conn)
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
			Owner: ADDR_01_LIKE,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
		{
			Id:     "nftlike1aaaaaa2",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
		{
			Id:     "nftlike1bbbbbbb",
			Parent: NftClassParent{IscnIdPrefix: prefixB},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-12908931",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-12908932",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_COSMOS,
		},
		{
			NftId:   "testing-nft-12908933",
			ClassId: nftClasses[1].Id,
			Owner:   ADDR_02_LIKE,
		},
		{
			NftId:   "testing-nft-12908934",
			ClassId: nftClasses[1].Id,
			Owner:   ADDR_02_COSMOS,
		},
		{
			NftId:   "testing-nft-12908935",
			ClassId: nftClasses[2].Id,
			Owner:   ADDR_03_LIKE,
		},
		{
			NftId:   "testing-nft-12908936",
			ClassId: nftClasses[2].Id,
			Owner:   ADDR_03_COSMOS,
		},
	}
	err := InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name  string
		query QueryNftCountRequest
		count uint64
	}{
		{"empty request", QueryNftCountRequest{}, 4},
		{
			"query with IncludeOwner = true",
			QueryNftCountRequest{IncludeOwner: true}, 6,
		},
		{
			"query with ignore list (1)", QueryNftCountRequest{
				IgnoreList: []string{ADDR_01_LIKE}, IncludeOwner: true,
			}, 4,
		},
		{
			"query with ignore list (1, 2)",
			QueryNftCountRequest{
				IgnoreList:   []string{ADDR_01_LIKE, ADDR_02_COSMOS},
				IncludeOwner: true,
			}, 2,
		},
	}
	for i, testCase := range testCases {
		res, err := GetNftCount(Conn, testCase.query)
		if err != nil {
			t.Errorf("test case #%02d (%s): GetNftCount returned error: %#v", i, testCase.name, err)
			continue
		}
		if res.Count != testCase.count {
			t.Errorf("test case #%02d (%s): expect count = %d, got %d.", i, testCase.name, testCase.count, res.Count)
			continue
		}
	}
}

func TestNftTradeStats(t *testing.T) {
	defer CleanupTestData(Conn)
	nftEvents := []NftEvent{
		{
			ClassId: "likenft1class1",
			NftId:   "testing-nft-1209301",
			Action:  ACTION_MINT,
			TxHash:  "A1",
		},
		{
			ClassId:  "likenft1class1",
			NftId:    "testing-nft-1209301",
			Action:   ACTION_SEND,
			Sender:   ADDR_01_LIKE,
			Receiver: ADDR_02_COSMOS,
			TxHash:   "A2",
		},
		{
			ClassId: "likenft1class2",
			NftId:   "testing-nft-1209302",
			Action:  ACTION_MINT,
			TxHash:  "B1",
		},
		{
			ClassId:  "likenft1class2",
			NftId:    "testing-nft-1209302",
			Action:   ACTION_SEND,
			Sender:   ADDR_01_LIKE,
			Receiver: ADDR_03_LIKE,
			TxHash:   "B2",
		},
		{
			ClassId: "likenft1class3",
			NftId:   "testing-nft-1209303",
			Action:  ACTION_MINT,
			TxHash:  "C1",
		},
		{
			ClassId:  "likenft1class3",
			NftId:    "testing-nft-1209303",
			Action:   ACTION_SEND,
			Sender:   ADDR_01_LIKE,
			Receiver: ADDR_04_LIKE,
			TxHash:   "C2",
		},
	}
	txs := []string{
		`{"txhash":"A1","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"1111"}]}]}]}}}`,
		`{"txhash":"A2","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"1000"}]}]}]}}}`,
		`{"txhash":"B1","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"2222"}]}]}]}}}`,
		`{"txhash":"B2","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"2000"}]}]}]}}}`,
		`{"txhash":"C1","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"3333"}]}]}]}}}`,
		`{"txhash":"C2","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"3000"}]}]}]}}}`,
	}
	err := InsertTestData(DBTestData{NftEvents: nftEvents, Txs: txs})
	if err != nil {
		t.Fatal(err)
	}

	query := QueryNftTradeStatsRequest{
		ApiAddresses: []string{ADDR_01_LIKE},
	}
	res, err := GetNftTradeStats(Conn, query)
	if err != nil {
		t.Fatal(err)
	}
	if res.Count != 3 {
		t.Fatalf("expect count = 3, got %d. result = %#v", res.Count, res)
	}
	if res.TotalVolume != 6000 {
		t.Fatalf("expect total volume = 6000, got %d. result = %#v", res.TotalVolume, res)
	}
}

func TestNftCreatorCount(t *testing.T) {
	defer CleanupTestData(Conn)
	nftEvents := []NftEvent{
		{
			ClassId: "likenft1class1",
			Action:  ACTION_NEW_CLASS,
			Sender:  ADDR_01_LIKE,
		},
		{
			ClassId: "likenft1class1",
			NftId:   "testing-nft-1209301",
			Action:  ACTION_MINT,
			Sender:  ADDR_02_LIKE,
		},
		{
			ClassId:  "likenft1class1",
			NftId:    "testing-nft-1209301",
			Action:   ACTION_SEND,
			Sender:   ADDR_03_LIKE,
			Receiver: ADDR_04_LIKE,
		},
		{
			ClassId: "likenft1class2",
			Action:  ACTION_NEW_CLASS,
			Sender:  ADDR_01_LIKE,
		},
		{
			ClassId: "likenft1class2",
			NftId:   "testing-nft-1209302",
			Action:  ACTION_MINT,
			Sender:  ADDR_05_LIKE,
		},
		{
			ClassId:  "likenft1class2",
			NftId:    "testing-nft-1209302",
			Action:   ACTION_SEND,
			Sender:   ADDR_06_LIKE,
			Receiver: ADDR_07_LIKE,
		},
		{
			ClassId: "likenft1class3",
			Action:  ACTION_NEW_CLASS,
			Sender:  ADDR_02_LIKE,
		},
		{
			ClassId: "likenft1class3",
			NftId:   "testing-nft-1209303",
			Action:  ACTION_MINT,
			Sender:  ADDR_05_LIKE,
		},
		{
			ClassId: "likenft1class4",
			Action:  ACTION_NEW_CLASS,
			Sender:  ADDR_03_LIKE,
		},
	}
	err := InsertTestData(DBTestData{NftEvents: nftEvents})
	if err != nil {
		t.Fatal(err)
	}

	res, err := GetNftCreatorCount(Conn)
	if err != nil {
		t.Fatal(err)
	}
	if res.Count != 3 {
		t.Fatalf("expect count = 3, got %d. result = %#v", res.Count, res)
	}
}

func TestNftOwnerCount(t *testing.T) {
	defer CleanupTestData(Conn)
	nfts := []Nft{
		{
			NftId: "testing-nft-1123123098",
			Owner: ADDR_01_LIKE,
		},
		{
			NftId: "testing-nft-1123123099",
			Owner: ADDR_01_LIKE,
		},
		{
			NftId: "testing-nft-1123123100",
			Owner: ADDR_01_LIKE,
		},
		{
			NftId: "testing-nft-1123123101",
			Owner: ADDR_02_LIKE,
		},
		{
			NftId: "testing-nft-1123123102",
			Owner: ADDR_02_LIKE,
		},
		{
			NftId: "testing-nft-1123123103",
			Owner: ADDR_03_LIKE,
		},
	}
	err := InsertTestData(DBTestData{Nfts: nfts})
	if err != nil {
		t.Fatal(err)
	}

	res, err := GetNftOwnerCount(Conn)
	if err != nil {
		t.Fatal(err)
	}
	if res.Count != 3 {
		t.Fatalf("expect count = 3, got %d. result = %#v", res.Count, res)
	}
}

func TestNftOwnerList(t *testing.T) {
	defer CleanupTestData(Conn)
	nfts := []Nft{
		{
			NftId: "testing-nft-1123123098",
			Owner: ADDR_01_LIKE,
		},
		{
			NftId: "testing-nft-1123123099",
			Owner: ADDR_01_LIKE,
		},
		{
			NftId: "testing-nft-1123123100",
			Owner: ADDR_01_LIKE,
		},
		{
			NftId: "testing-nft-1123123101",
			Owner: ADDR_02_LIKE,
		},
		{
			NftId: "testing-nft-1123123102",
			Owner: ADDR_02_LIKE,
		},
		{
			NftId: "testing-nft-1123123103",
			Owner: ADDR_03_LIKE,
		},
	}
	err := InsertTestData(DBTestData{Nfts: nfts})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name       string
		pagination PageRequest
		owners     []string
		counts     []int
	}{
		{
			"limit = 100",
			PageRequest{Limit: 100},
			[]string{ADDR_01_LIKE, ADDR_02_LIKE, ADDR_03_LIKE},
			[]int{3, 2, 1},
		},
		{
			"limit = 1",
			PageRequest{Limit: 1},
			[]string{ADDR_01_LIKE},
			[]int{3},
		},
		{
			"limit = 2, offset = 1",
			PageRequest{Limit: 2, Offset: 1},
			[]string{ADDR_02_LIKE, ADDR_03_LIKE},
			[]int{2, 1},
		},
	}

NEXT_TESTCASE:
	for i, testCase := range testCases {
		res, err := GetNftOwnerList(Conn, testCase.pagination)
		if err != nil {
			t.Errorf("test case #%02d: GetNftOwnerList returned error: %#v", i, err)
			continue NEXT_TESTCASE
		}
		if len(res.Owners) != len(testCase.owners) {
			t.Errorf("test case #%02d: expect owners count = %d, got %d. results = %#v", i, len(testCase.owners), len(res.Owners), res.Owners)
			continue NEXT_TESTCASE
		}
		for j, resOwner := range res.Owners {
			if resOwner.Owner != testCase.owners[j] {
				t.Errorf("test case #%02d: expect owner = %s, got %s. results = %#v", i, testCase.owners[j], resOwner.Owner, res.Owners)
				continue NEXT_TESTCASE
			}
			if resOwner.Count != testCase.counts[j] {
				t.Errorf("test case #%02d: expect count for owner %s = %d, got %d. results = %#v", i, testCase.owners[j], testCase.counts[j], resOwner.Count, res.Owners)
				continue NEXT_TESTCASE
			}
		}
	}
}
