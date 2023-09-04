package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestISCNRecordCount(t *testing.T) {
	defer CleanupTestData(Conn)
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
	InsertTestData(DBTestData{Iscns: iscns})

	res, err := GetISCNRecordCount(Conn)
	require.NoError(t, err)
	require.Equal(t, uint64(2), res.Count)
}

func TestISCNOwnerCount(t *testing.T) {
	defer CleanupTestData(Conn)
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
	InsertTestData(DBTestData{Iscns: iscns})

	res, err := GetISCNOwnerCount(Conn)
	require.NoError(t, err)
	require.Equal(t, uint64(2), res.Count)
}

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
	InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts})

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
		require.NoError(t, err, "test case #%02d (%s): GetNftCount returned error: %#v", i, testCase.name, err)
		require.Equal(t, testCase.count, res.Count, "test case #%02d (%s): expect count = %d, got %d.", i, testCase.name, testCase.count, res.Count)
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
			Price:    10,
		},
		{
			ClassId:  "likenft1class1",
			NftId:    "testing-nft-1209301",
			Action:   ACTION_SEND,
			Sender:   ADDR_01_LIKE,
			Receiver: ADDR_02_COSMOS,
			TxHash:   "A3",
			Price:    0,
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
			Price:    20,
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
			Price:    30,
		},
		{
			ClassId:  "likenft1class3",
			NftId:    "testing-nft-1209303",
			Action:   ACTION_BUY,
			Sender:   ADDR_01_LIKE,
			Receiver: ADDR_05_LIKE,
			TxHash:   "C3",
			Price:    40,
		},
		{
			ClassId:  "likenft1class3",
			NftId:    "testing-nft-1209303",
			Action:   ACTION_SELL,
			Sender:   ADDR_01_LIKE,
			Receiver: ADDR_06_LIKE,
			TxHash:   "C4",
			Price:    50,
		},
	}
	InsertTestData(DBTestData{NftEvents: nftEvents})

	query := QueryNftTradeStatsRequest{}
	res, err := GetNftTradeStats(Conn, query)
	require.NoError(t, err)
	require.Equal(t, QueryNftTradeStatsResponse{
		Count:       5,
		TotalVolume: 150,
	}, res)
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
	InsertTestData(DBTestData{NftEvents: nftEvents})

	res, err := GetNftCreatorCount(Conn)
	require.NoError(t, err)
	require.Equal(t, uint64(3), res.Count)
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
	InsertTestData(DBTestData{Nfts: nfts})

	res, err := GetNftOwnerCount(Conn)
	require.NoError(t, err)
	require.Equal(t, uint64(3), res.Count)
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
	InsertTestData(DBTestData{Nfts: nfts})

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

	for i, testCase := range testCases {
		res, err := GetNftOwnerList(Conn, testCase.pagination)
		require.NoError(t, err)
		require.Equal(t, len(testCase.owners), len(res.Owners), "test case #%02d: %s", i, testCase.name)
		for j, resOwner := range res.Owners {
			require.Equal(t, testCase.owners[j], resOwner.Owner, "test case #%02d: %s", i, testCase.name)
			require.Equal(t, testCase.counts[j], resOwner.Count, "test case #%02d: %s", i, testCase.name)
		}
	}
}
