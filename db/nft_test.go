package db_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func TestQueryNftClass(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	prefixB := "iscn://testing/bbbbbb"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_01_COSMOS,
		},
		{
			Iscn:  "iscn://testing/aaaaaa/2",
			Owner: ADDR_02_COSMOS,
		},
		{
			Iscn:  "iscn://testing/bbbbbb/1",
			Owner: ADDR_03_COSMOS,
		},
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaa1",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
		{
			Id:     "nftlike1aaaaa2",
			Parent: NftClassParent{IscnIdPrefix: prefixA},
		},
		{
			Id:     "nftlike1bbbbb1",
			Parent: NftClassParent{IscnIdPrefix: prefixB},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-1123123098",
			ClassId: nftClasses[0].Id,
		},
		{
			NftId:   "testing-nft-1123123099",
			ClassId: nftClasses[0].Id,
		},
		{
			NftId:   "testing-nft-1123123100",
			ClassId: nftClasses[1].Id,
		},
		{
			NftId:   "testing-nft-1123123101",
			ClassId: nftClasses[1].Id,
		},
		{
			NftId:   "testing-nft-1123123102",
			ClassId: nftClasses[2].Id,
		},
		{
			NftId:   "testing-nft-1123123103",
			ClassId: nftClasses[2].Id,
		},
	}
	InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts})

	testCases := []struct {
		name     string
		query    QueryClassRequest
		count    int
		classIds []string
	}{
		{"empty requset", QueryClassRequest{}, 3, nil},
		{
			"query by iscn prefix, AllIscnVersions = true",
			QueryClassRequest{
				IscnIdPrefix:    nftClasses[0].Parent.IscnIdPrefix,
				AllIscnVersions: true,
			}, 2, []string{nftClasses[0].Id, nftClasses[1].Id},
		},
		{
			"query by iscn prefix, AllIscnVersions = false (1)",
			QueryClassRequest{
				IscnIdPrefix:    nftClasses[0].Parent.IscnIdPrefix,
				AllIscnVersions: false,
			}, 2, []string{nftClasses[0].Id, nftClasses[1].Id},
		},
		{
			"query by iscn prefix, AllIscnVersions = false (2)",
			QueryClassRequest{
				IscnIdPrefix: nftClasses[2].Parent.IscnIdPrefix,
			}, 1, []string{nftClasses[2].Id},
		},
		{
			"query by iscn owner 1 (LIKE prefix), AllIscnVersions = true",
			QueryClassRequest{
				IscnOwner:       []string{ADDR_01_LIKE},
				AllIscnVersions: true,
			}, 2, []string{nftClasses[0].Id, nftClasses[1].Id},
		},
		{
			"query by iscn owner 1 (LIKE prefix), AllIscnVersions = false",
			QueryClassRequest{
				IscnOwner:       []string{ADDR_01_LIKE},
				AllIscnVersions: false,
			}, 0, nil,
		},
		{
			"query by iscn owner 2 (LIKE prefix), AllIscnVersions = true",
			QueryClassRequest{
				IscnOwner:       []string{ADDR_02_LIKE},
				AllIscnVersions: true,
			}, 2, []string{nftClasses[0].Id, nftClasses[1].Id},
		},
		{
			"query by iscn owner 3 (LIKE prefix), AllIscnVersions = false",
			QueryClassRequest{IscnOwner: []string{ADDR_03_LIKE}}, 1,
			[]string{nftClasses[2].Id},
		},
		{
			"query by non existing iscn prefix",
			QueryClassRequest{IscnIdPrefix: "nosuchprefix"}, 0, nil,
		},
		{
			"query by non existing iscn owner",
			QueryClassRequest{IscnOwner: []string{"nosuchowner"}}, 0, nil,
		},
		{
			"query by multiple iscn owners",
			QueryClassRequest{
				IscnOwner: []string{ADDR_02_LIKE, ADDR_03_LIKE},
			}, 3, []string{nftClasses[0].Id, nftClasses[1].Id, nftClasses[2].Id},
		},
	}

	p := PageRequest{
		Limit: 10,
	}
NEXT_TESTCASE:
	for i, testCase := range testCases {
		res, err := GetClasses(Conn, testCase.query, p)
		require.NoError(t, err, "Error in test case #%02d (%s)", i, testCase.name)
		require.Len(t, res.Classes, testCase.count, "error in test case #%02d (%s)", i, testCase.name)
		for _, c := range res.Classes {
			for _, n := range c.Nfts {
				require.Equal(t, c.Id, n.ClassId, "error in test case #%02d (%s)", i, testCase.name)
			}
		}
		if len(testCase.classIds) > 0 {
		NEXT_CLASS:
			for _, c := range res.Classes {
				for _, id := range testCase.classIds {
					if c.Id == id {
						continue NEXT_CLASS
					}
				}
				t.Errorf("test case #%02d (%s): expect classId in %#v, got %s. results = %#v", i, testCase.name, testCase.classIds, c.Id, res.Classes)
				continue NEXT_TESTCASE
			}
		}
	}
}

func TestQueryNftByOwner(t *testing.T) {
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
	}
	nftClasses := []NftClass{
		{
			Id:     "nftlike1aaaaaa",
			Parent: NftClassParent{IscnIdPrefix: "iscn://testing/aaaaaa"},
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-1123123098",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_03_LIKE,
		},
	}
	nftEvents := []NftEvent{
		{
			ClassId:   nfts[0].ClassId,
			NftId:     nfts[0].NftId,
			Action:    ACTION_SEND,
			Sender:    ADDR_01_LIKE,
			Receiver:  nfts[0].Owner,
			TxHash:    "AA",
			Timestamp: time.Unix(1, 0),
		},
	}
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		NftEvents:  nftEvents,
	})

	testCases := []struct {
		owner string
		count int
	}{
		{nfts[0].Owner, 1},
		{ADDR_03_COSMOS, 1},
		{ADDR_02_COSMOS, 0},
	}

	p := PageRequest{
		Limit: 10,
	}
	for i, testCase := range testCases {
		query := QueryNftRequest{Owner: testCase.owner}
		res, err := GetNfts(Conn, query, p)
		require.NoError(t, err, "Error in test case #%02d (owner = %s, ExpandClasses = false)", i, testCase.owner)
		require.Len(t, res.Nfts, testCase.count, "error in test case #%02d (owner = %s, ExpandClasses = false)", i, testCase.owner)
		for _, n := range res.Nfts {
			require.Nil(t, n.ClassData, "error in test case #%02d (owner = %s, ExpandClasses = false)", i, testCase.owner)
		}
		query.ExpandClasses = true
		res, err = GetNfts(Conn, query, p)
		require.NoError(t, err, "Error in test case #%02d (owner = %s, ExpandClasses = true)", i, testCase.owner)
		require.Len(t, res.Nfts, testCase.count, "error in test case #%02d (owner = %s, ExpandClasses = true)", i, testCase.owner)
		for _, n := range res.Nfts {
			require.NotNil(t, n.ClassData, "error in test case #%02d (owner = %s, ExpandClasses = true)", i, testCase.owner)
			require.Equal(t, n.ClassId, n.ClassData.Id, "error in test case #%02d (owner = %s, ExpandClasses = true)", i, testCase.owner)
		}
	}
}

func TestOwnerByClassId(t *testing.T) {
	defer CleanupTestData(Conn)
	nfts := []Nft{
		{
			NftId:   "testing-nft-1123123098",
			ClassId: "likenft1xxxxxx",
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-1123123099",
			ClassId: "likenft1xxxxxx",
			Owner:   ADDR_02_LIKE,
		},
		{
			NftId:   "testing-nft-1123123100",
			ClassId: "likenft1xxxxxx",
			Owner:   ADDR_02_LIKE,
		},
		{
			NftId:   "testing-nft-1123123101",
			ClassId: "likenft1yyyyyy",
			Owner:   ADDR_02_LIKE,
		},
		{
			NftId:   "testing-nft-1123123102",
			ClassId: "likenft1yyyyyy",
			Owner:   ADDR_03_LIKE,
		},
		{
			NftId:   "testing-nft-1123123103",
			ClassId: "likenft1zzzzzz",
			Owner:   ADDR_03_LIKE,
		},
	}
	InsertTestData(DBTestData{Nfts: nfts})

	testCases := []struct {
		classId string
		owners  []string
		counts  []int
	}{
		{
			nfts[0].ClassId,
			[]string{ADDR_01_LIKE, ADDR_02_LIKE},
			[]int{1, 2},
		},
		{
			nfts[3].ClassId,
			[]string{ADDR_02_LIKE, ADDR_03_LIKE},
			[]int{1, 1},
		},
		{
			nfts[5].ClassId,
			[]string{ADDR_03_LIKE},
			[]int{1},
		},
		{"likenft1notexist", nil, nil},
	}

	for i, testCase := range testCases {
		query := QueryOwnerRequest{ClassId: testCase.classId}
		res, err := GetOwners(Conn, query)
		require.NoError(t, err, "Error in test case #%02d (classId = %s)", i, testCase.classId)
		require.Len(t, res.Owners, len(testCase.owners), "error in test case #%02d (classId = %s)", i, testCase.classId)
	NEXT_OWNER:
		for j, owner := range testCase.owners {
			if len(testCase.owners) == 0 {
				continue NEXT_OWNER
			}
			for _, ownerRes := range res.Owners {
				if ownerRes.Owner == owner {
					require.Equal(t, testCase.counts[j], ownerRes.Count, "error in test case #%02d (classId = %s)", i, testCase.classId)
					continue NEXT_OWNER
				}
			}
			t.Errorf("test case #%02d (classId %s): owner %s not found. results = %#v", i, testCase.classId, owner, res.Owners)
		}
	}
}

func TestNftEvents(t *testing.T) {
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
	nftClasses := []NftClass{
		{
			Id:     "likenft1aaaaaa",
			Parent: NftClassParent{IscnIdPrefix: "iscn://testing/aaaaaa"},
		},
		{
			Id:     "likenft1bbbbbb",
			Parent: NftClassParent{IscnIdPrefix: "iscn://testing/bbbbbb"},
		},
	}
	nftEvents := []NftEvent{
		{
			ClassId:  nftClasses[0].Id,
			NftId:    "testing-nft-12093810",
			Action:   "action1",
			Sender:   ADDR_01_LIKE,
			Receiver: ADDR_02_COSMOS,
			TxHash:   "AAAAAA",
		},
		{
			ClassId:  nftClasses[0].Id,
			NftId:    "testing-nft-12093811",
			Action:   "action2",
			Sender:   ADDR_02_COSMOS,
			Receiver: ADDR_03_LIKE,
			TxHash:   "BBBBBB",
		},
		{
			ClassId:  nftClasses[1].Id,
			NftId:    "testing-nft-12093812",
			Action:   "action2",
			Sender:   ADDR_03_LIKE,
			Receiver: ADDR_01_COSMOS,
			TxHash:   "CCCCCC",
		},
	}
	txs := []string{
		// hack: set all memo equal to txhash for easier checking
		`{"txhash":"AAAAAA","tx":{"body":{"memo":"AAAAAA"}}}`,
		`{"txhash":"BBBBBB","tx":{"body":{"memo":"BBBBBB"}}}`,
		`{"txhash":"CCCCCC","tx":{"body":{"memo":"CCCCCC"}}}`,
	}
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		NftEvents:  nftEvents,
		Txs:        txs,
	})

	testCases := []struct {
		name    string
		query   QueryEventsRequest
		count   int
		checker func(int, []NftEvent)
	}{
		{
			"empty request", QueryEventsRequest{}, 3,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.TxHash, e.Memo, "error in test case #%02d: expect memo = %s, got %s. events = %#v", i, e.TxHash, e.Memo, events)
				}
			},
		},
		{
			"query by class ID (0)", QueryEventsRequest{ClassId: nftClasses[0].Id}, 2,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.ClassId, nftClasses[0].Id, "error in test case #%02d: expect classId = %s, got %s. events = %#v", i, nftClasses[0].Id, e.ClassId, events)
				}
			},
		},
		{
			"query by class ID (1)", QueryEventsRequest{ClassId: nftClasses[1].Id}, 1,
			func(i int, events []NftEvent) {
				e := events[0]
				require.Equal(t, e.ClassId, nftClasses[1].Id, "error in test case #%02d: expect classId = %s, got %s. events = %#v", i, nftClasses[1].Id, e.ClassId, events)
			},
		},
		{
			"query by class ID (0) and NFT ID (0)", QueryEventsRequest{
				ClassId: nftClasses[0].Id, NftId: nftEvents[0].NftId,
			},
			1,
			func(i int, events []NftEvent) {
				e := events[0]
				require.Equal(t, e.ClassId, nftClasses[0].Id, "error in test case #%02d: expect classId = %s, got %s. events = %#v", i, nftClasses[0].Id, e.ClassId, events)
				require.Equal(t, e.NftId, nftEvents[0].NftId, "error in test case #%02d: expect nftId = %s, got %s. events = %#v", i, nftEvents[0].NftId, e.NftId, events)
			},
		},
		{
			"query by action type (0)",
			QueryEventsRequest{
				ActionType: []NftEventAction{nftEvents[0].Action},
			},
			1,
			func(i int, events []NftEvent) {
				e := events[0]
				require.Equal(t, e.Action, nftEvents[0].Action, "error in test case #%02d: expect action = %s, got %s. events = %#v", i, nftEvents[0].Action, e.Action, events)
			},
		},
		{
			"query by action type (1)",
			QueryEventsRequest{
				ActionType: []NftEventAction{nftEvents[1].Action},
			},
			2,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.Action, nftEvents[1].Action, "error in test case #%02d: expect action = %s, got %s. events = %#v", i, nftEvents[1].Action, e.Action, events)
				}
			},
		},
		{
			"query by action type (0, 1)",
			QueryEventsRequest{
				ActionType: []NftEventAction{nftEvents[0].Action, nftEvents[1].Action},
			},
			3,
			nil,
		},
		{
			"query with ignore from list (1)",
			QueryEventsRequest{IgnoreFromList: []string{ADDR_01_COSMOS}},
			2,
			func(i int, events []NftEvent) {
				ignoredList := utils.ConvertAddressPrefixes(ADDR_01_COSMOS, AddressPrefixes)
				for _, e := range events {
					for _, ignoredAddr := range ignoredList {
						require.NotEqual(t, e.Sender, ignoredAddr, "error in test case #%02d: expect sender ignored, got sender = %s. events = %#v", i, e.Sender, events)
					}
				}
			},
		},
		{
			"query with ignore from list (1, 2)",
			QueryEventsRequest{IgnoreFromList: []string{ADDR_01_COSMOS, ADDR_02_LIKE}},
			1,
			func(i int, events []NftEvent) {
				ignoredList := utils.ConvertAddressArrayPrefixes([]string{ADDR_01_COSMOS, ADDR_02_LIKE}, AddressPrefixes)
				for _, e := range events {
					for _, ignoredAddr := range ignoredList {
						require.NotEqual(t, e.Sender, ignoredAddr, "error in test case #%02d: expect sender ignored, got sender = %s. events = %#v", i, e.Sender, events)
					}
				}
			},
		},
		{
			"query with ignore to list (1)",
			QueryEventsRequest{IgnoreToList: []string{ADDR_01_COSMOS}},
			2,
			func(i int, events []NftEvent) {
				ignoredList := utils.ConvertAddressPrefixes(ADDR_01_COSMOS, AddressPrefixes)
				for _, e := range events {
					for _, ignoredAddr := range ignoredList {
						require.NotEqual(t, e.Receiver, ignoredAddr, "error in test case #%02d: expect receiver ignored, got receiver = %s. events = %#v", i, e.Receiver, events)
					}
				}
			},
		},
		{
			"query with ignore to list (1, 2)",
			QueryEventsRequest{IgnoreToList: []string{ADDR_01_COSMOS, ADDR_02_LIKE}},
			1,
			func(i int, events []NftEvent) {
				ignoredList := utils.ConvertAddressArrayPrefixes([]string{ADDR_01_COSMOS, ADDR_02_LIKE}, AddressPrefixes)
				for _, e := range events {
					for _, ignoredAddr := range ignoredList {
						require.NotEqual(t, e.Receiver, ignoredAddr, "error in test case #%02d: expect receiver ignored, got receiver = %s. events = %#v", i, e.Receiver, events)
					}
				}
			},
		},
		{
			"query by iscn ID prefix (0)",
			QueryEventsRequest{IscnIdPrefix: nftClasses[0].Parent.IscnIdPrefix},
			2,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.ClassId, nftClasses[0].Id, "error in test case #%02d: expect classId = %s (with ISCN prefix %s), got classId = %s. events = %#v", i, nftClasses[0].Id, nftClasses[0].Parent.IscnIdPrefix, e.ClassId, events)
				}
			},
		},
		{
			"query by iscn ID prefix (1)",
			QueryEventsRequest{IscnIdPrefix: nftClasses[1].Parent.IscnIdPrefix},
			1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.ClassId, nftClasses[1].Id, "error in test case #%02d: expect classId = %s (with ISCN prefix %s), got classId = %s. events = %#v", i, nftClasses[1].Id, nftClasses[1].Parent.IscnIdPrefix, e.ClassId, events)
				}
			},
		},
		{
			"query by sender with like prefix (0)", QueryEventsRequest{Sender: []string{ADDR_01_LIKE}}, 1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.Sender, ADDR_01_LIKE, "error in test case #%02d: expect sender = %s, got sender = %s. events = %#v", i, ADDR_01_LIKE, e.Sender, events)
				}
			},
		},
		{
			"query by sender with cosmos prefix (1)", QueryEventsRequest{Sender: []string{ADDR_02_COSMOS}}, 1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.Sender, ADDR_02_LIKE, "error in test case #%02d: expect sender = %s, got sender = %s. events = %#v", i, ADDR_02_LIKE, e.Sender, events)
				}
			},
		},
		{
			"query by receiver with cosmos prefix (0)", QueryEventsRequest{Receiver: []string{ADDR_02_COSMOS}}, 1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.Receiver, ADDR_02_LIKE, "error in test case #%02d: expect receiver = %s, got receiver = %s. events = %#v", i, ADDR_02_LIKE, e.Receiver, events)
				}
			},
		},
		{
			"query by receiver with like prefix (1)", QueryEventsRequest{Receiver: []string{ADDR_03_LIKE}}, 1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.Receiver, ADDR_03_LIKE, "error in test case #%02d: expect receiver = %s, got receiver = %s. events = %#v", i, ADDR_03_LIKE, e.Receiver, events)
				}
			},
		},
		{
			"query by creator", QueryEventsRequest{Creator: []string{iscns[1].Owner}}, 2,
			func(i int, events []NftEvent) {
				for _, e := range events {
					require.Equal(t, e.ClassId, nftClasses[0].Id, "error in test case #%02d: expect classId = %s (with ISCN Prefix %s), got classId = %s. events = %#v", i, nftClasses[0].Id, nftClasses[0].Parent.IscnIdPrefix, e.ClassId, events)
				}
			},
		},
		{
			"query by involver (1)", QueryEventsRequest{Involver: []string{ADDR_01_LIKE}}, 2, nil,
		},
		{
			"query by involver (2)", QueryEventsRequest{Involver: []string{ADDR_02_LIKE}}, 2, nil,
		},
		{
			"query by involver (3)", QueryEventsRequest{Involver: []string{ADDR_03_LIKE}}, 2, nil,
		},
		{
			"query by multiple senders (0, 1)", QueryEventsRequest{Sender: []string{ADDR_01_LIKE, ADDR_02_COSMOS}}, 2, nil,
		},
		{
			"query by multiple receivers (0, 1)", QueryEventsRequest{Receiver: []string{ADDR_02_COSMOS, ADDR_03_LIKE}}, 2, nil,
		},
		{
			"query by multiple creators (0, 1)", QueryEventsRequest{Creator: []string{iscns[0].Owner, iscns[1].Owner}}, 2, nil,
		},
		{
			"query by multiple creators (1, 2)", QueryEventsRequest{Creator: []string{iscns[1].Owner, iscns[2].Owner}}, 3, nil,
		},
		{"query by old creator", QueryEventsRequest{Creator: []string{ADDR_01_LIKE}}, 0, nil},
		{"query by non existing iscn ID prefix", QueryEventsRequest{IscnIdPrefix: "iscn://testing/notexist"}, 0, nil},
		{"query by non existing class ID", QueryEventsRequest{ClassId: "likenft1notexist"}, 0, nil},
		{"query by non existing NFT ID", QueryEventsRequest{ClassId: nftClasses[0].Id, NftId: "notexist"}, 0, nil},
		{"query by non existing sender", QueryEventsRequest{Sender: []string{ADDR_04_LIKE}}, 0, nil},
		{"query by non existing receiver", QueryEventsRequest{Receiver: []string{ADDR_04_LIKE}}, 0, nil},
	}

	p := PageRequest{
		Limit: 10,
	}
	for i, testCase := range testCases {
		res, err := GetNftEvents(Conn, testCase.query, p)
		require.NoError(t, err, "test case #%02d (%s): GetNftEvents returned error: %#v", i, testCase.name, err)
		require.Len(t, res.Events, testCase.count, "test case #%02d (%s): expect len(res.Events) = %d, got %d. results = %#v", i, testCase.name, testCase.count, len(res.Events), res.Events)
		if testCase.checker != nil {
			testCase.checker(i, res.Events)
		}
	}
}

func TestQueryNftRanking(t *testing.T) {
	defer CleanupTestData(Conn)
	prefixA := "iscn://testing/aaaaaa"
	prefixB := "iscn://testing/bbbbbb"
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: ADDR_05_LIKE,
			Data:  []byte(`{"contentMetadata":{"@type":"CreativeWorks"}}`),
			Stakeholders: []Stakeholder{
				{Entity: Entity{Id: "stakeholder_id_a1", Name: "stakeholder_name_a1"}},
				{Entity: Entity{Id: "stakeholder_id_a2", Name: "stakeholder_name_a2"}},
			},
		},
		{
			Iscn:  "iscn://testing/aaaaaa/2",
			Owner: ADDR_06_LIKE,
			Data:  []byte(`{"contentMetadata":{"@type":"CreativeWorks"}}`),
			Stakeholders: []Stakeholder{
				{Entity: Entity{Id: "stakeholder_id_a1", Name: "stakeholder_name_a1"}},
				{Entity: Entity{Id: "stakeholder_id_a2", Name: "stakeholder_name_a2"}},
			},
		},
		{
			Iscn:  "iscn://testing/bbbbbb/1",
			Owner: ADDR_07_LIKE,
			Data:  []byte(`{"contentMetadata":{"@type":"Article"}}`),
			Stakeholders: []Stakeholder{
				{Entity: Entity{Id: "stakeholder_id_b1", Name: "stakeholder_name_b1"}},
				{Entity: Entity{Id: "stakeholder_id_b2", Name: "stakeholder_name_b2"}},
			},
			Timestamp: time.Unix(2, 0),
		},
	}
	nftClasses := []NftClass{
		{
			Id:        "likenft1aaaaaa",
			Parent:    NftClassParent{IscnIdPrefix: prefixA},
			CreatedAt: time.Unix(1, 0),
		},
		{
			Id:        "likenft1bbbbbb",
			Parent:    NftClassParent{IscnIdPrefix: prefixB},
			CreatedAt: time.Unix(2, 0),
		},
	}
	nfts := []Nft{
		{
			ClassId: nftClasses[0].Id,
			NftId:   "testing-nft-12093810",
			Owner:   ADDR_02_COSMOS,
		},
		{
			ClassId: nftClasses[0].Id,
			NftId:   "testing-nft-12093811",
			Owner:   ADDR_03_LIKE,
		},
		{
			ClassId: nftClasses[1].Id,
			NftId:   "testing-nft-12093812",
			Owner:   ADDR_03_LIKE,
		},
	}
	nftEvents := []NftEvent{
		{
			ClassId:   nfts[0].ClassId,
			NftId:     nfts[0].NftId,
			Action:    ACTION_MINT,
			TxHash:    "A1",
			Timestamp: time.Unix(1, 0),
			Price:     1111,
		},
		{
			ClassId:   nfts[0].ClassId,
			NftId:     nfts[0].NftId,
			Action:    ACTION_SEND,
			Sender:    ADDR_01_LIKE,
			Receiver:  nfts[0].Owner,
			TxHash:    "A2",
			Timestamp: time.Unix(2, 0),
			Price:     1000,
		},
		{
			ClassId:   nfts[1].ClassId,
			NftId:     nfts[1].NftId,
			Action:    ACTION_MINT,
			TxHash:    "B1",
			Timestamp: time.Unix(3, 0),
			Price:     2222,
		},
		{
			ClassId:   nfts[1].ClassId,
			NftId:     nfts[1].NftId,
			Action:    ACTION_SEND,
			Sender:    ADDR_01_LIKE,
			Receiver:  nfts[1].Owner,
			TxHash:    "B2",
			Timestamp: time.Unix(4, 0),
			Price:     2000,
		},
		{
			ClassId:   nfts[2].ClassId,
			NftId:     nfts[2].NftId,
			Action:    ACTION_MINT,
			TxHash:    "C1",
			Timestamp: time.Unix(5, 0),
			Price:     3333,
		},
		{
			ClassId:   nfts[2].ClassId,
			NftId:     nfts[2].NftId,
			Action:    ACTION_SEND,
			Sender:    ADDR_01_LIKE,
			Receiver:  nfts[2].Owner,
			TxHash:    "C2",
			Timestamp: time.Unix(6, 0),
			Price:     2500,
		},
	}
	InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		NftEvents:  nftEvents,
	})

	testCases := []struct {
		name            string
		query           QueryRankingRequest
		classIDs        []string
		totalSoldValues []int64
		soldCounts      []int
	}{
		{
			name:            "empty request",
			query:           QueryRankingRequest{},
			classIDs:        []string{nftClasses[0].Id, nftClasses[1].Id},
			soldCounts:      []int{2, 1},
			totalSoldValues: []int64{3000, 2500},
		},
		{
			name: "query by creator (1) in old version of iscn",
			// ADDR_01_LIKE is owner of old version, so should return no result
			query: QueryRankingRequest{Creator: ADDR_01_LIKE},
		},
		{
			name:            "query by creator (2)",
			query:           QueryRankingRequest{Creator: ADDR_06_COSMOS},
			classIDs:        []string{nftClasses[0].Id},
			soldCounts:      []int{2},
			totalSoldValues: []int64{3000},
		},
		{
			name:            "query by type",
			query:           QueryRankingRequest{Type: "CreativeWorks"},
			classIDs:        []string{nftClasses[0].Id},
			soldCounts:      []int{2},
			totalSoldValues: []int64{3000},
		},
		{
			name:            "query by stakeholder ID",
			query:           QueryRankingRequest{StakeholderId: iscns[0].Stakeholders[0].Entity.Id},
			classIDs:        []string{nftClasses[0].Id},
			soldCounts:      []int{2},
			totalSoldValues: []int64{3000},
		},
		{
			name:            "query by stakeholder name",
			query:           QueryRankingRequest{StakeholderName: iscns[2].Stakeholders[1].Entity.Name},
			classIDs:        []string{nftClasses[1].Id},
			soldCounts:      []int{1},
			totalSoldValues: []int64{2500},
		},
		{
			name:            "query by collector",
			query:           QueryRankingRequest{Collector: ADDR_03_COSMOS},
			classIDs:        []string{nftClasses[1].Id, nftClasses[0].Id},
			soldCounts:      []int{1, 1},
			totalSoldValues: []int64{2500, 2000},
		},
		{
			name:            "query created after",
			query:           QueryRankingRequest{CreatedAfter: 1},
			classIDs:        []string{nftClasses[1].Id},
			soldCounts:      []int{1},
			totalSoldValues: []int64{2500},
		},
		{
			name:            "query created before",
			query:           QueryRankingRequest{CreatedBefore: 2},
			classIDs:        []string{nftClasses[0].Id},
			soldCounts:      []int{2},
			totalSoldValues: []int64{3000},
		},
		{
			name:            "query sold after",
			query:           QueryRankingRequest{After: 4},
			classIDs:        []string{nftClasses[1].Id},
			soldCounts:      []int{1},
			totalSoldValues: []int64{2500},
		},
		{
			name:            "query sold before",
			query:           QueryRankingRequest{Before: 4},
			classIDs:        []string{nftClasses[0].Id},
			soldCounts:      []int{1},
			totalSoldValues: []int64{1000},
		},
	}

	p := PageRequest{
		Limit: 100,
	}

	for i, testCase := range testCases {
		testCase.query.ApiAddresses = []string{ADDR_01_LIKE}
		res, err := GetClassesRanking(Conn, testCase.query, p)
		require.NoError(t, err, "test case #%02d (%s): GetClassesRanking returned error: %#v", i, testCase.name, err)
		require.Len(t, res.Classes, len(testCase.classIDs), "test case #%02d (%s): expect len(res.Classes) = %d, got %d. results = %#v", i, testCase.name, len(testCase.classIDs), len(res.Classes), res.Classes)
		for j, class := range res.Classes {
			require.Equal(t, testCase.classIDs[j], class.NftClass.Id, "test case #%02d (%s), class %d: expect class ID = %s, got %s. results = %#v", i, testCase.name, j, testCase.classIDs[j], class.NftClass.Id, res.Classes)
			require.Equal(t, testCase.soldCounts[j], class.SoldCount, "test case #%02d (%s), class %d: expect sold count = %d, got %d. results = %#v", i, testCase.name, j, testCase.soldCounts[j], class.SoldCount, res.Classes)
			require.Equal(t, testCase.totalSoldValues[j], class.TotalSoldValue, "test case #%02d (%s), class %d: expect total sold value = %d, got %d. results = %#v", i, testCase.name, j, testCase.totalSoldValues[j], class.TotalSoldValue, res.Classes)
		}
	}
}
func TestCollectors(t *testing.T) {
	defer CleanupTestData(Conn)
	creators := []string{ADDR_01_LIKE, ADDR_02_LIKE, ADDR_03_LIKE, ADDR_04_LIKE}
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: creators[0],
		},
		{
			Iscn:  "iscn://testing/aaaaaa/2",
			Owner: creators[1],
		},
		{
			Iscn:  "iscn://testing/bbbbbb/1",
			Owner: creators[2],
		},
		{
			Iscn:  "iscn://testing/cccccc/1",
			Owner: creators[3],
		},
	}
	nftClasses := []NftClass{
		{Id: "likenft1aaaaaa", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/aaaaaa"}, LatestPrice: 100},
		{Id: "likenft1bbbbbb", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/bbbbbb"}, LatestPrice: 1000},
		{Id: "likenft1cccccc", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/cccccc"}, LatestPrice: 10000},
	}
	collectors := []string{ADDR_02_LIKE, ADDR_03_LIKE, ADDR_05_LIKE, ADDR_06_LIKE}
	nfts := []Nft{
		{
			ClassId:     nftClasses[0].Id,
			NftId:       "testing-nft-12093810",
			Owner:       collectors[0],
			LatestPrice: 100,
		},
		{
			ClassId:     nftClasses[0].Id,
			NftId:       "testing-nft-12093811",
			Owner:       collectors[1],
			LatestPrice: 10,
		},
		{
			ClassId:     nftClasses[1].Id,
			NftId:       "testing-nft-12093812",
			Owner:       collectors[1],
			LatestPrice: 20,
		},
		{
			ClassId:     nftClasses[0].Id,
			NftId:       "testing-nft-12093813",
			Owner:       collectors[2],
			LatestPrice: nftClasses[0].LatestPrice,
		},
		{
			ClassId:     nftClasses[1].Id,
			NftId:       "testing-nft-12093814",
			Owner:       collectors[2],
			LatestPrice: nftClasses[1].LatestPrice,
		},
		{
			ClassId:     nftClasses[2].Id,
			NftId:       "testing-nft-12093815",
			Owner:       collectors[2],
			LatestPrice: nftClasses[2].LatestPrice,
		},
	}
	nftEvents := []NftEvent{
		{
			ClassId:  nfts[0].ClassId,
			NftId:    nfts[0].NftId,
			Receiver: nfts[0].Owner,
			Price:    nfts[0].LatestPrice,
		},
		{
			ClassId:  nfts[1].ClassId,
			NftId:    nfts[1].NftId,
			Receiver: nfts[1].Owner,
			Price:    nfts[1].LatestPrice,
		},
		{
			ClassId:  nfts[2].ClassId,
			NftId:    nfts[2].NftId,
			Receiver: nfts[2].Owner,
			Price:    nfts[2].LatestPrice,
		},
		{
			ClassId:  nfts[3].ClassId,
			NftId:    nfts[3].NftId,
			Receiver: nfts[3].Owner,
			Price:    nfts[3].LatestPrice,
		},
		{
			ClassId:  nfts[4].ClassId,
			NftId:    nfts[4].NftId,
			Receiver: nfts[4].Owner,
			Price:    nfts[4].LatestPrice,
		},
		// add action and txHash to comply with table constraint
		// sender is for explanatory purposes only
		{
			Action:   ACTION_SEND,
			ClassId:  nfts[5].ClassId,
			NftId:    nfts[5].NftId,
			Receiver: nfts[5].Owner,
			TxHash:   "AAAAAA",
			Price:    1,
		},
		{
			Action:   ACTION_SELL,
			ClassId:  nfts[5].ClassId,
			NftId:    nfts[5].NftId,
			Sender:   nfts[5].Owner,
			Receiver: collectors[3],
			TxHash:   "BBBBBB",
			Price:    2,
		},
		{
			Action:   ACTION_BUY,
			ClassId:  nfts[5].ClassId,
			NftId:    nfts[5].NftId,
			Sender:   collectors[3],
			Receiver: nfts[5].Owner,
			TxHash:   "CCCCCC",
			Price:    nfts[5].LatestPrice,
		},
	}
	InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts, NftEvents: nftEvents})

	// end state:
	// collectors[0]: 1 NFT, NFT price = 100, class price = 100, from himself
	// collectors[1]: 2 NFTs, NFT price = 30, class price = 1100, 1 from himself, 1 from other
	// collectors[2]: 3 NFTs, NFT price = 11100, class price = 11100, all from others
	// collectors[3]: 0 NFT, bought 1 NFT from collectors[2], and sold it back

	testCases := []struct {
		name        string
		query       QueryCollectorRequest
		owners      []string
		totalValues []uint64
	}{
		// HACK: default IncludeOwner is true when binding to form
		{
			name:   "empty query",
			query:  QueryCollectorRequest{IncludeOwner: true},
			owners: []string{collectors[2], collectors[0], collectors[1]},
			totalValues: []uint64{
				nfts[3].LatestPrice + nfts[4].LatestPrice + nfts[5].LatestPrice,
				nfts[0].LatestPrice,
				nfts[1].LatestPrice + nfts[2].LatestPrice,
			},
		},
		{
			name:        "query with IncludeOwner = false",
			query:       QueryCollectorRequest{IncludeOwner: false},
			owners:      []string{collectors[1], collectors[2]},
			totalValues: []uint64{nfts[1].LatestPrice, nfts[3].LatestPrice + nfts[4].LatestPrice + nfts[5].LatestPrice},
		},
		{
			name:        "query by creator, AllIscnVersions = false",
			query:       QueryCollectorRequest{Creator: creators[0], IncludeOwner: true},
			owners:      []string{},
			totalValues: []uint64{},
		},
		{
			name:        "query by creator, AllIscnVersions = true",
			query:       QueryCollectorRequest{Creator: creators[0], IncludeOwner: true, AllIscnVersions: true},
			owners:      []string{collectors[0], collectors[2], collectors[1]},
			totalValues: []uint64{nfts[0].LatestPrice, nfts[3].LatestPrice, nfts[1].LatestPrice},
		},
		{
			name: "query with ignore list",
			query: QueryCollectorRequest{
				IgnoreList:   []string{collectors[1], collectors[2]},
				IncludeOwner: true,
			},
			owners:      []string{collectors[0]},
			totalValues: []uint64{nfts[0].LatestPrice},
		},
		{
			name:   "query with PriceBy = class",
			query:  QueryCollectorRequest{IncludeOwner: true, PriceBy: "class"},
			owners: []string{collectors[2], collectors[0], collectors[1]},
			totalValues: []uint64{
				nftClasses[0].LatestPrice + nftClasses[1].LatestPrice + nftClasses[2].LatestPrice,
				nftClasses[0].LatestPrice,
				nftClasses[0].LatestPrice + nftClasses[1].LatestPrice,
			},
		},
		{
			name:   "query with OrderBy = count",
			query:  QueryCollectorRequest{IncludeOwner: true, OrderBy: "count"},
			owners: []string{collectors[2], collectors[1], collectors[0]},
			totalValues: []uint64{
				nfts[3].LatestPrice + nfts[4].LatestPrice + nfts[5].LatestPrice,
				nfts[1].LatestPrice + nfts[2].LatestPrice,
				nfts[0].LatestPrice,
			},
		},
	}

	p := PageRequest{
		Limit: 10,
	}
	for i, testCase := range testCases {
		res, err := GetCollector(Conn, testCase.query, p)
		require.NoError(t, err, "test case #%02d (%s): GetCollector returned error: %#v", i, testCase.name, err)
		require.Len(t, res.Collectors, len(testCase.owners), "test case #%02d (%s): expect len(res.Collectors) = %d, got %d. results = %#v", i, testCase.name, len(testCase.owners), len(res.Collectors), res.Collectors)
		if len(res.Collectors) > 1 {
			for j := 1; j < len(res.Collectors); j++ {
				prev := res.Collectors[j-1]
				curr := res.Collectors[j]
				if testCase.query.OrderBy == "count" {
					require.LessOrEqual(t, curr.Count, prev.Count, "test case #%02d (%s): expect Collectors in descending order, got results = %#v", i, testCase.name, res.Collectors)
				} else {
					require.LessOrEqual(t, curr.TotalValue, prev.TotalValue, "test case #%02d (%s): expect Collectors in descending order, got results = %#v", i, testCase.name, res.Collectors)
				}
			}
		}
	NEXT_OWNER:
		for j, owner := range testCase.owners {
			for _, collector := range res.Collectors {
				if collector.Account == owner {
					require.Equal(t, testCase.totalValues[j], collector.TotalValue, "test case #%02d (%s), collector %s: expect total value = %d, got %d. results = %#v", i, testCase.name, owner, testCase.totalValues[j], collector.TotalValue, res.Collectors)
					continue NEXT_OWNER
				}
			}
			t.Errorf("test case #%02d (%s): collector %s not found. results = %#v", i, testCase.name, owner, res.Collectors)
		}
	}
}

func TestCreators(t *testing.T) {
	defer CleanupTestData(Conn)
	creators := []string{ADDR_01_LIKE, ADDR_02_LIKE, ADDR_03_LIKE, ADDR_04_LIKE}
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/aaaaaa/1",
			Owner: creators[0],
		},
		{
			Iscn:  "iscn://testing/aaaaaa/2",
			Owner: creators[1],
		},
		{
			Iscn:  "iscn://testing/bbbbbb/1",
			Owner: creators[2],
		},
		{
			Iscn:  "iscn://testing/cccccc/1",
			Owner: creators[3],
		},
	}
	nftClasses := []NftClass{
		{Id: "likenft1aaaaaa", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/aaaaaa"}, LatestPrice: 100},
		{Id: "likenft1bbbbbb", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/bbbbbb"}, LatestPrice: 1000},
		{Id: "likenft1cccccc", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/cccccc"}, LatestPrice: 10000},
	}
	collectors := []string{ADDR_02_LIKE, ADDR_03_LIKE, ADDR_05_LIKE, ADDR_06_LIKE}
	nfts := []Nft{
		{
			ClassId:     nftClasses[0].Id,
			NftId:       "testing-nft-12093810",
			Owner:       collectors[0],
			LatestPrice: 100,
		},
		{
			ClassId:     nftClasses[0].Id,
			NftId:       "testing-nft-12093811",
			Owner:       collectors[1],
			LatestPrice: 10,
		},
		{
			ClassId:     nftClasses[1].Id,
			NftId:       "testing-nft-12093812",
			Owner:       collectors[1],
			LatestPrice: 20,
		},
		{
			ClassId:     nftClasses[0].Id,
			NftId:       "testing-nft-12093813",
			Owner:       collectors[2],
			LatestPrice: nftClasses[0].LatestPrice,
		},
		{
			ClassId:     nftClasses[1].Id,
			NftId:       "testing-nft-12093814",
			Owner:       collectors[2],
			LatestPrice: nftClasses[1].LatestPrice,
		},
		{
			ClassId:     nftClasses[2].Id,
			NftId:       "testing-nft-12093815",
			Owner:       collectors[2],
			LatestPrice: nftClasses[2].LatestPrice,
		},
	}
	nftEvents := []NftEvent{
		{
			ClassId:  nfts[0].ClassId,
			NftId:    nfts[0].NftId,
			Receiver: nfts[0].Owner,
			Price:    nfts[0].LatestPrice,
		},
		{
			ClassId:  nfts[1].ClassId,
			NftId:    nfts[1].NftId,
			Receiver: nfts[1].Owner,
			Price:    nfts[1].LatestPrice,
		},
		{
			ClassId:  nfts[2].ClassId,
			NftId:    nfts[2].NftId,
			Receiver: nfts[2].Owner,
			Price:    nfts[2].LatestPrice,
		},
		{
			ClassId:  nfts[3].ClassId,
			NftId:    nfts[3].NftId,
			Receiver: nfts[3].Owner,
			Price:    nfts[3].LatestPrice,
		},
		{
			ClassId:  nfts[4].ClassId,
			NftId:    nfts[4].NftId,
			Receiver: nfts[4].Owner,
			Price:    nfts[4].LatestPrice,
		},
		// add action and txHash to comply with table constraint
		// sender is for explanatory purposes only
		{
			Action:   ACTION_SEND,
			ClassId:  nfts[5].ClassId,
			NftId:    nfts[5].NftId,
			Receiver: nfts[5].Owner,
			TxHash:   "AAAAAA",
			Price:    1,
		},
		{
			Action:   ACTION_SELL,
			ClassId:  nfts[5].ClassId,
			NftId:    nfts[5].NftId,
			Sender:   nfts[5].Owner,
			Receiver: collectors[3],
			TxHash:   "BBBBBB",
			Price:    2,
		},
		{
			Action:   ACTION_BUY,
			ClassId:  nfts[5].ClassId,
			NftId:    nfts[5].NftId,
			Sender:   collectors[3],
			Receiver: nfts[5].Owner,
			TxHash:   "CCCCCC",
			Price:    nfts[5].LatestPrice,
		},
	}
	InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts, NftEvents: nftEvents})

	// end state:
	// creators[0]: transfer ISCN ownership to collectors[1]
	// creators[1]: 3 NFTs sold, NFT price = 210, class price = 300
	// creators[2]: 2 NFTs sold, NFT price = 1020, class price = 2000
	// creators[3]: 1 NFT sold, NFT price = 10000, class price = 10000

	testCases := []struct {
		name        string
		query       QueryCreatorRequest
		owners      []string
		totalValues []uint64
	}{
		// HACK: default IncludeOwner is true when binding to form
		{
			name:   "empty query",
			query:  QueryCreatorRequest{IncludeOwner: true},
			owners: []string{creators[3], creators[2], creators[1]},
			totalValues: []uint64{
				nfts[5].LatestPrice,
				nfts[2].LatestPrice + nfts[4].LatestPrice,
				nfts[0].LatestPrice + nfts[1].LatestPrice + nfts[3].LatestPrice,
			},
		},
		{
			name:        "query by collector (0), AllIscnVersions = false",
			query:       QueryCreatorRequest{Collector: collectors[0], IncludeOwner: true},
			owners:      []string{creators[1]},
			totalValues: []uint64{nfts[0].LatestPrice},
		},
		{
			name:        "query by collector (0), AllIscnVersions = false, IncludeOwner = false",
			query:       QueryCreatorRequest{Collector: collectors[0], IncludeOwner: false},
			owners:      []string{},
			totalValues: []uint64{},
		},
		{
			name:        "query by collector (1), AllIscnVersions = false",
			query:       QueryCreatorRequest{Collector: collectors[1], IncludeOwner: true},
			owners:      []string{creators[2], creators[1]},
			totalValues: []uint64{nfts[2].LatestPrice, nfts[1].LatestPrice},
		},
		{
			name:        "query by collector (1), AllIscnVersions = true",
			query:       QueryCreatorRequest{Collector: collectors[1], IncludeOwner: true, AllIscnVersions: true},
			owners:      []string{creators[2], creators[1], creators[0]},
			totalValues: []uint64{nfts[2].LatestPrice, nfts[1].LatestPrice, nfts[1].LatestPrice},
		},
		{
			name:   "AllIscnVersions = false, PriceBy = class",
			query:  QueryCreatorRequest{IncludeOwner: true, PriceBy: "class"},
			owners: []string{creators[3], creators[2], creators[1]},
			totalValues: []uint64{
				nftClasses[2].LatestPrice,
				nftClasses[1].LatestPrice * 2,
				nftClasses[0].LatestPrice * 3,
			},
		},
		{
			name:   "AllIscnVersions = false, OrderBy = count",
			query:  QueryCreatorRequest{IncludeOwner: true, OrderBy: "count"},
			owners: []string{creators[1], creators[2], creators[3]},
			totalValues: []uint64{
				nfts[0].LatestPrice + nfts[1].LatestPrice + nfts[3].LatestPrice,
				nfts[2].LatestPrice + nfts[4].LatestPrice,
				nfts[5].LatestPrice,
			},
		},
	}

	p := PageRequest{
		Limit: 10,
	}
	for i, testCase := range testCases {
		res, err := GetCreators(Conn, testCase.query, p)
		require.NoError(t, err, "test case #%02d (%s)", i, testCase.name)
		require.Len(t, res.Creators, len(testCase.owners), "test case #%02d (%s)", i, testCase.name)
		if len(res.Creators) > 1 {
			for j := 1; j < len(res.Creators); j++ {
				prev := res.Creators[j-1]
				curr := res.Creators[j]
				if testCase.query.OrderBy == "count" {
					require.LessOrEqual(t, curr.Count, prev.Count, "test case #%02d (%s)", i, testCase.name)
				} else {
					require.LessOrEqual(t, curr.TotalValue, prev.TotalValue, "test case #%02d (%s)", i, testCase.name)
				}
			}
		}
	NEXT_OWNER:
		for j, owner := range testCase.owners {
			for _, creator := range res.Creators {
				if creator.Account == owner {
					require.Equal(t, testCase.owners[j], creator.Account, "test case #%02d (%s)", i, testCase.name)
					continue NEXT_OWNER
				}
			}
			t.Errorf("test case #%02d (%s): creator %s not found. results = %#v", i, testCase.name, owner, res.Creators)
		}
	}
}

func TestGetCollectorTopRankedCreators(t *testing.T) {
	defer CleanupTestData(Conn)
	r := rand.New(rand.NewSource(19823981948123019))
	iscns := []IscnInsert{}
	nftClasses := []NftClass{}
	nfts := []Nft{}
	nftEvents := []NftEvent{}
	for iscnOwnerIndex := 0; iscnOwnerIndex < 5; iscnOwnerIndex++ {
		for iscnIndex := 0; iscnIndex < 5; iscnIndex++ {
			iscnPrefix := fmt.Sprintf("iscn://testing/%02d-%02d", iscnOwnerIndex, iscnIndex)
			iscns = append(iscns, IscnInsert{
				Iscn:  iscnPrefix + "/1",
				Owner: ADDRS_LIKE[iscnOwnerIndex],
			})
			for nftClassIndex := 0; nftClassIndex < 5; nftClassIndex++ {
				classId := fmt.Sprintf("testing-nft-class-%s-%02d", iscnPrefix, nftClassIndex)
				nftClasses = append(nftClasses, NftClass{
					Id:     classId,
					Parent: NftClassParent{IscnIdPrefix: iscnPrefix},
				})
				for nftIndex := 0; nftIndex < 10; nftIndex++ {
					nftId := fmt.Sprintf("%s-%02d", classId, nftIndex)
					owner := ADDRS_LIKE[r.Intn(10)]
					price := uint64(r.Int31n(2) + 1)
					nfts = append(nfts, Nft{
						ClassId:     classId,
						NftId:       nftId,
						Owner:       owner,
						LatestPrice: price,
					})
					nftEvents = append(nftEvents, NftEvent{
						ClassId:  classId,
						NftId:    nftId,
						Receiver: owner,
						Price:    price,
					})
				}
			}
		}
	}
	InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts, NftEvents: nftEvents})

	p := PageRequest{
		Limit:   100,
		Reverse: true,
	}

	shouldBeRankK := func(t *testing.T, res QueryCollectorResponse, collector string, k uint) {
		collectorTotalValue := uint64(0)
		for _, collectorEntry := range res.Collectors {
			if collectorEntry.Account == collector {
				collectorTotalValue = collectorEntry.TotalValue
				break
			}
		}
		require.NotZero(t, collectorTotalValue)

		inFrontOfCollectorCount := uint(0)
		for _, collectorEntry := range res.Collectors {
			if collectorEntry.TotalValue > collectorTotalValue {
				inFrontOfCollectorCount++
			}
		}
		require.Equal(t, k-1, inFrontOfCollectorCount)
	}

	for i := 0; i < 10; i++ {
		for k := uint(1); k <= 10; k++ {
			collector := ADDRS_LIKE[i]
			res, err := GetCollectorTopRankedCreators(Conn, QueryCollectorTopRankedCreatorsRequest{
				Collector:    collector,
				IncludeOwner: true,
				Top:          k,
			})
			require.NoError(t, err)

			for _, creatorEntry := range res.Creators {
				require.LessOrEqual(t, creatorEntry.Rank, k)
				res, err := GetCollector(Conn, QueryCollectorRequest{
					Creator:      creatorEntry.Creator,
					IncludeOwner: true,
				}, p)
				require.NoError(t, err)
				require.NotEmpty(t, res.Collectors)
				shouldBeRankK(t, res, collector, creatorEntry.Rank)
			}
		}
	}
}

func TestQueryClassesOwned(t *testing.T) {
	defer CleanupTestData(Conn)
	nftClasses := []NftClass{
		{
			Id: "nftlike1aaaaa1",
		},
		{
			Id: "nftlike1bbbbb1",
		},
		{
			Id: "nftlike1ccccc1",
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-109283748",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-109283749",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_02_LIKE,
		},
		{
			NftId:   "testing-nft-109283750",
			ClassId: nftClasses[1].Id,
			Owner:   ADDR_03_LIKE,
		},
		{
			NftId:   "testing-nft-109283751",
			ClassId: nftClasses[1].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-109283752",
			ClassId: nftClasses[2].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-109283753",
			ClassId: nftClasses[2].Id,
			Owner:   ADDR_02_LIKE,
		},
	}
	InsertTestData(DBTestData{NftClasses: nftClasses, Nfts: nfts})

	testCases := []struct {
		name     string
		query    QueryClassesOwnedRequest
		classIds []string
	}{
		{
			"query owner 01, class IDs (0, 1, 2)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id, nftClasses[1].Id, nftClasses[2].Id},
				Owner:    ADDR_01_LIKE,
			}, []string{nftClasses[0].Id, nftClasses[1].Id, nftClasses[2].Id},
		},
		{
			"query owner 02, class IDs (0, 1, 2)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id, nftClasses[1].Id, nftClasses[2].Id},
				Owner:    ADDR_02_LIKE,
			}, []string{nftClasses[0].Id, nftClasses[2].Id},
		},
		{
			"query owner 03, class IDs (0, 1, 2)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id, nftClasses[1].Id, nftClasses[2].Id},
				Owner:    ADDR_03_LIKE,
			}, []string{nftClasses[1].Id},
		},
		{
			"query owner 01, class IDs (0)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id},
				Owner:    ADDR_01_LIKE,
			}, []string{nftClasses[0].Id},
		},
		{
			"query owner 02, class IDs (0)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id},
				Owner:    ADDR_02_LIKE,
			}, []string{nftClasses[0].Id},
		},
		{
			"query owner 03, class IDs (0)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id},
				Owner:    ADDR_03_LIKE,
			}, []string{},
		},
		{
			"query owner 01, class IDs (0, 1)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id, nftClasses[1].Id},
				Owner:    ADDR_01_LIKE,
			}, []string{nftClasses[0].Id, nftClasses[1].Id},
		},
		{
			"query owner 02, class IDs (0, 1)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id, nftClasses[1].Id},
				Owner:    ADDR_02_LIKE,
			}, []string{nftClasses[0].Id},
		},
		{
			"query owner 03, class IDs (0, 1)",
			QueryClassesOwnedRequest{
				ClassIds: []string{nftClasses[0].Id, nftClasses[1].Id},
				Owner:    ADDR_03_LIKE,
			}, []string{nftClasses[1].Id},
		},
	}

	for i, testCase := range testCases {
		res, err := GetClassesOwned(Conn, testCase.query)
		require.NoError(t, err, "Error in test case #%02d (%s)", i, testCase.name)
		require.ElementsMatch(t, res.ClassIds, testCase.classIds, "Error in test case #%02d (%s)", i, testCase.name)
	}
}
