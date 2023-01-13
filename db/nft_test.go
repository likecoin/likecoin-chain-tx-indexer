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
	err := InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts})
	if err != nil {
		t.Fatal(err)
	}

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
				IscnOwner:       ADDR_01_LIKE,
				AllIscnVersions: true,
			}, 2, []string{nftClasses[0].Id, nftClasses[1].Id},
		},
		{
			"query by iscn owner 1 (LIKE prefix), AllIscnVersions = false",
			QueryClassRequest{
				IscnOwner:       ADDR_01_LIKE,
				AllIscnVersions: false,
			}, 0, nil,
		},
		{
			"query by iscn owner 2 (LIKE prefix), AllIscnVersions = true",
			QueryClassRequest{
				IscnOwner:       ADDR_02_LIKE,
				AllIscnVersions: true,
			}, 2, []string{nftClasses[0].Id, nftClasses[1].Id},
		},
		{
			"query by iscn owner 3 (LIKE prefix), AllIscnVersions = false",
			QueryClassRequest{IscnOwner: ADDR_03_LIKE}, 1,
			[]string{nftClasses[2].Id},
		},
		{
			"query by non existing iscn prefix",
			QueryClassRequest{IscnIdPrefix: "nosuchprefix"}, 0, nil,
		},
		{
			"query by non existing iscn owner",
			QueryClassRequest{IscnOwner: "nosuchowner"}, 0, nil,
		},
	}

	p := PageRequest{
		Limit: 10,
	}
NEXT_TESTCASE:
	for i, testCase := range testCases {
		res, err := GetClasses(Conn, testCase.query, p)
		if err != nil {
			t.Errorf("test case #%02d (%s): GetClasses returned error: %#v", i, testCase.name, err)
			continue NEXT_TESTCASE
		}
		if len(res.Classes) != testCase.count {
			t.Errorf("test case #%02d (%s): expect len(res.Classes) = %d, got %d. results = %#v", i, testCase.name, testCase.count, len(res.Classes), res.Classes)
			continue NEXT_TESTCASE
		}
		for _, c := range res.Classes {
			for _, n := range c.Nfts {
				if n.ClassId != c.Id {
					t.Errorf("test case #%02d (%s): expect all nft under nft class has same classId as the class (%s), got %s. results = %#v", i, testCase.name, c.Id, n.ClassId, res.Classes)
					continue NEXT_TESTCASE
				}
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
			Action:    "/cosmos.nft.v1beta1.MsgSend",
			Sender:    ADDR_01_LIKE,
			Receiver:  nfts[0].Owner,
			TxHash:    "AA",
			Timestamp: time.Unix(1, 0),
		},
	}
	err := InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		NftEvents:  nftEvents,
	})
	if err != nil {
		t.Fatal(err)
	}

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
NEXT_TESTCASE:
	for i, testCase := range testCases {
		query := QueryNftRequest{Owner: testCase.owner}
		res, err := GetNfts(Conn, query, p)
		if err != nil {
			t.Errorf("test case #%02d (owner = %s, ExpandClasses = false): GetNfts returned error: %#v", i, testCase.owner, err)
			continue NEXT_TESTCASE
		}
		if len(res.Nfts) != testCase.count {
			t.Errorf("test case #%02d (owner = %s, ExpandClasses = false): expect len(res.Nfts) = %d, got %d. results = %#v", i, testCase.owner, testCase.count, len(res.Nfts), res.Nfts)
			continue NEXT_TESTCASE
		}
		for _, n := range res.Nfts {
			if n.ClassData != nil {
				t.Errorf("test case #%02d (owner = %s, ExpandClasses = false): ClassData should be nil. results = %#v", i, testCase.owner, res.Nfts)
				continue NEXT_TESTCASE
			}
		}
		query.ExpandClasses = true
		res, err = GetNfts(Conn, query, p)
		if err != nil {
			t.Errorf("test case #%02d (owner = %s, ExpandClasses = true): GetNfts returned error: %#v", i, testCase.owner, err)
			continue NEXT_TESTCASE
		}
		if len(res.Nfts) != testCase.count {
			t.Errorf("test case #%02d (owner = %s, ExpandClasses = true): expect len(res.Nfts) = %d, got %d. results = %#v", i, testCase.owner, testCase.count, len(res.Nfts), res.Nfts)
			continue NEXT_TESTCASE
		}
		for _, n := range res.Nfts {
			if n.ClassData == nil {
				t.Errorf("test case #%02d (owner = %s, ExpandClasses = true): ClassData should not be nil. results = %#v", i, testCase.owner, res.Nfts)
				continue NEXT_TESTCASE
			}
			if n.ClassData.Id != n.ClassId {
				t.Errorf("test case #%02d (owner = %s, ExpandClasses = true): NFT class ID not equal to the ID in expanded class data. results = %#v", i, testCase.owner, res.Nfts)
				continue NEXT_TESTCASE
			}
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
	err := InsertTestData(DBTestData{Nfts: nfts})
	if err != nil {
		t.Fatal(err)
	}

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

NEXT_TESTCASE:
	for i, testCase := range testCases {
		query := QueryOwnerRequest{ClassId: testCase.classId}
		res, err := GetOwners(Conn, query)
		if err != nil {
			t.Errorf("test case #%02d (classId = %s): GetOwners returned error: %#v", i, testCase.classId, err)
			continue NEXT_TESTCASE
		}
		if len(res.Owners) != len(testCase.owners) {
			t.Errorf("test case #%02d (classId = %s): expect len(res.Owners) = %d, got %d. results = %#v", i, testCase.classId, len(testCase.owners), len(res.Owners), res.Owners)
			continue NEXT_TESTCASE
		}
	NEXT_OWNER:
		for j, owner := range testCase.owners {
			if len(testCase.owners) == 0 {
				continue NEXT_OWNER
			}
			for _, ownerRes := range res.Owners {
				if ownerRes.Owner == owner {
					if ownerRes.Count != testCase.counts[j] {
						t.Errorf("test case #%02d (classId = %s), owner %s: expect count = %d, got %d. results = %#v", i, testCase.classId, owner, testCase.counts[j], ownerRes.Count, res.Owners)
						continue NEXT_TESTCASE
					}
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
	err := InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		NftEvents:  nftEvents,
		Txs:        txs,
	})
	if err != nil {
		t.Fatal(err)
	}

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
					if e.Memo != e.TxHash {
						t.Errorf(`test case #%02d: expect memo = %s, got %s. events = %#v`, i, e.TxHash, e.Memo, events)
						return
					}
				}
			},
		},
		{
			"query by class ID (0)", QueryEventsRequest{ClassId: nftClasses[0].Id}, 2,
			func(i int, events []NftEvent) {
				for _, e := range events {
					if e.ClassId != nftClasses[0].Id {
						t.Errorf(`test case #%02d: expect classId = %s, got %s. events = %#v`, i, nftClasses[0].Id, e.ClassId, events)
						return
					}
				}
			},
		},
		{
			"query by class ID (1)", QueryEventsRequest{ClassId: nftClasses[1].Id}, 1,
			func(i int, events []NftEvent) {
				e := events[0]
				if e.ClassId != nftClasses[1].Id {
					t.Errorf(`test case #%02d: expect classId = %s, got %s. events = %#v`, i, nftClasses[0].Id, e.ClassId, events)
				}
			},
		},
		{
			"query by class ID (0) and NFT ID (0)", QueryEventsRequest{
				ClassId: nftClasses[0].Id, NftId: nftEvents[0].NftId,
			},
			1,
			func(i int, events []NftEvent) {
				e := events[0]
				if e.ClassId != nftClasses[0].Id {
					t.Errorf(`test case #%02d: expect classId = %s, got %s. events = %#v`, i, nftClasses[0].Id, e.ClassId, events)
				} else if e.NftId != nftEvents[0].NftId {
					t.Errorf(`test case #%02d: expect nftId = %s, got %s. events = %#v`, i, nftEvents[0].NftId, e.NftId, events)
				}
			},
		},
		{
			"query by action type (0)",
			QueryEventsRequest{
				ActionType: []string{nftEvents[0].Action},
			},
			1,
			func(i int, events []NftEvent) {
				e := events[0]
				if e.Action != nftEvents[0].Action {
					t.Errorf(`test case #%02d: expect action = %s, got %s. events = %#v`, i, nftEvents[0].Action, e.Action, events)
				}
			},
		},
		{
			"query by action type (1)",
			QueryEventsRequest{
				ActionType: []string{nftEvents[1].Action},
			},
			2,
			func(i int, events []NftEvent) {
				for _, e := range events {
					if e.Action != nftEvents[1].Action {
						t.Errorf(`test case #%02d: expect action = %s, got %s. events = %#v`, i, nftEvents[1].Action, e.Action, events)
						return
					}
				}
			},
		},
		{
			"query by action type (0, 1)",
			QueryEventsRequest{
				ActionType: []string{nftEvents[0].Action, nftEvents[1].Action},
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
						if e.Sender == ignoredAddr {
							t.Errorf(`test case #%02d: expect sender ignored, got sender = %s. events = %#v`, i, e.Sender, events)
							return
						}
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
						if e.Sender == ignoredAddr {
							t.Errorf(`test case #%02d: expect sender ignored, got sender = %s. events = %#v`, i, e.Sender, events)
							return
						}
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
						if e.Receiver == ignoredAddr {
							t.Errorf(`test case #%02d: expect receiver ignored, got receiver = %s. events = %#v`, i, e.Receiver, events)
							return
						}
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
						if e.Receiver == ignoredAddr {
							t.Errorf(`test case #%02d: expect receiver ignored, got receiver = %s. events = %#v`, i, e.Receiver, events)
							return
						}
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
					if e.ClassId != nftClasses[0].Id {
						t.Errorf(`test case #%02d: expect classId = %s (with ISCN prefix %s), got classId = %s. events = %#v`, i, nftClasses[0].Id, nftClasses[0].Parent.IscnIdPrefix, e.ClassId, events)
						return
					}
				}
			},
		},
		{
			"query by iscn ID prefix (1)",
			QueryEventsRequest{IscnIdPrefix: nftClasses[1].Parent.IscnIdPrefix},
			1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					if e.ClassId != nftClasses[1].Id {
						t.Errorf(`test case #%02d: expect classId = %s (with ISCN prefix %s), got classId = %s. events = %#v`, i, nftClasses[1].Id, nftClasses[1].Parent.IscnIdPrefix, e.ClassId, events)
						return
					}
				}
			},
		},
		{
			"query by sender with like prefix (0)", QueryEventsRequest{Sender: ADDR_01_LIKE}, 1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					if e.Sender != ADDR_01_LIKE {
						t.Errorf(`test case #%02d: expect sender = %s, got sender = %s. events = %#v`, i, ADDR_01_LIKE, e.Sender, events)
						return
					}
				}
			},
		},
		{
			"query by sender with cosmos prefix (1)", QueryEventsRequest{Sender: ADDR_02_COSMOS}, 1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					if e.Sender != ADDR_02_LIKE {
						t.Errorf(`test case #%02d: expect sender = %s, got sender = %s. events = %#v`, i, ADDR_02_LIKE, e.Sender, events)
						return
					}
				}
			},
		},
		{
			"query by receiver with cosmos prefix (0)", QueryEventsRequest{Receiver: ADDR_02_COSMOS}, 1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					if e.Receiver != ADDR_02_LIKE {
						t.Errorf(`test case #%02d: expect receiver = %s, got receiver = %s. events = %#v`, i, ADDR_02_LIKE, e.Receiver, events)
						return
					}
				}
			},
		},
		{
			"query by receiver with like prefix (1)", QueryEventsRequest{Receiver: ADDR_03_LIKE}, 1,
			func(i int, events []NftEvent) {
				for _, e := range events {
					if e.Receiver != ADDR_03_LIKE {
						t.Errorf(`test case #%02d: expect receiver = %s, got receiver = %s. events = %#v`, i, ADDR_03_LIKE, e.Receiver, events)
						return
					}
				}
			},
		},
		{
			"query by creator", QueryEventsRequest{Creator: iscns[1].Owner}, 2,
			func(i int, events []NftEvent) {
				for _, e := range events {
					if e.ClassId != nftClasses[0].Id {
						t.Errorf(`test case #%02d: expect classId = %s (with ISCN Prefix %s), got classId = %s. events = %#v`,
							i, nftClasses[0].Id, nftClasses[0].Parent.IscnIdPrefix, e.ClassId, events)
						return
					}
				}
			},
		},
		{"query by old creator", QueryEventsRequest{Creator: ADDR_01_LIKE}, 0, nil},
		{"query by non existing iscn ID prefix", QueryEventsRequest{IscnIdPrefix: "iscn://testing/notexist"}, 0, nil},
		{"query by non existing class ID", QueryEventsRequest{ClassId: "likenft1notexist"}, 0, nil},
		{"query by non existing NFT ID", QueryEventsRequest{ClassId: nftClasses[0].Id, NftId: "notexist"}, 0, nil},
		{"query by non existing sender", QueryEventsRequest{Sender: ADDR_04_LIKE}, 0, nil},
		{"query by non existing receiver", QueryEventsRequest{Receiver: ADDR_04_LIKE}, 0, nil},
	}

	p := PageRequest{
		Limit: 10,
	}
	for i, testCase := range testCases {
		res, err := GetNftEvents(Conn, testCase.query, p)
		if err != nil {
			t.Errorf("test case #%02d (%s): GetNftEvents returned error: %#v", i, testCase.name, err)
			continue
		}
		if len(res.Events) != testCase.count {
			t.Errorf("test case #%02d (%s): expect len(res.Events) = %d, got %d. results = %#v", i, testCase.name, testCase.count, len(res.Events), res.Events)
			continue
		}
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
			Action:    "mint_nft",
			TxHash:    "A1",
			Timestamp: time.Unix(1, 0),
		},
		{
			ClassId:   nfts[0].ClassId,
			NftId:     nfts[0].NftId,
			Action:    "/cosmos.nft.v1beta1.MsgSend",
			Sender:    ADDR_01_LIKE,
			Receiver:  nfts[0].Owner,
			TxHash:    "A2",
			Timestamp: time.Unix(2, 0),
		},
		{
			ClassId:   nfts[1].ClassId,
			NftId:     nfts[1].NftId,
			Action:    "mint_nft",
			TxHash:    "B1",
			Timestamp: time.Unix(3, 0),
		},
		{
			ClassId:   nfts[1].ClassId,
			NftId:     nfts[1].NftId,
			Action:    "/cosmos.nft.v1beta1.MsgSend",
			Sender:    ADDR_01_LIKE,
			Receiver:  nfts[1].Owner,
			TxHash:    "B2",
			Timestamp: time.Unix(4, 0),
		},
		{
			ClassId:   nfts[2].ClassId,
			NftId:     nfts[2].NftId,
			Action:    "mint_nft",
			TxHash:    "C1",
			Timestamp: time.Unix(5, 0),
		},
		{
			ClassId:   nfts[2].ClassId,
			NftId:     nfts[2].NftId,
			Action:    "/cosmos.nft.v1beta1.MsgSend",
			Sender:    ADDR_01_LIKE,
			Receiver:  nfts[2].Owner,
			TxHash:    "C2",
			Timestamp: time.Unix(6, 0),
		},
	}
	txs := []string{
		`{"txhash":"A1","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"1111"}]}]}]}}}`,
		`{"txhash":"A2","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"1000"}]}]}]}}}`,
		`{"txhash":"B1","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"2222"}]}]}]}}}`,
		`{"txhash":"B2","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"2000"}]}]}]}}}`,
		`{"txhash":"C1","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"3333"}]}]}]}}}`,
		`{"txhash":"C2","tx":{"body":{"messages":[{"msgs":[{"amount":[{"amount":"2500"}]}]}]}}}`,
	}
	err := InsertTestData(DBTestData{
		Iscns:      iscns,
		NftClasses: nftClasses,
		Nfts:       nfts,
		NftEvents:  nftEvents,
		Txs:        txs,
	})
	if err != nil {
		t.Fatal(err)
	}

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
		if err != nil {
			t.Errorf("test case #%02d (%s): GetClassesRanking returned error: %#v", i, testCase.name, err)
			continue
		}
		if len(res.Classes) != len(testCase.classIDs) {
			t.Errorf("test case #%02d (%s): expect len(res.Classes) = %d, got %d. results = %#v", i, testCase.name, len(testCase.classIDs), len(res.Classes), res.Classes)
			continue
		}
		for j, class := range res.Classes {
			if class.NftClass.Id != testCase.classIDs[j] {
				t.Errorf("test case #%02d (%s), class %d: expect class ID = %s, got %s. results = %#v", i, testCase.name, j, testCase.classIDs[j], class.NftClass.Id, res.Classes)
				continue
			}
			if class.SoldCount != testCase.soldCounts[j] {
				t.Errorf("test case #%02d (%s), class %d: expect sold count = %d, got %d. results = %#v", i, testCase.name, j, testCase.soldCounts[j], class.SoldCount, res.Classes)
				continue
			}
			if class.TotalSoldValue != testCase.totalSoldValues[j] {
				t.Errorf("test case #%02d (%s), class %d: expect total sold value = %d, got %d. results = %#v", i, testCase.name, j, testCase.totalSoldValues[j], class.TotalSoldValue, res.Classes)
				continue
			}
		}
	}
}
func TestCollectors(t *testing.T) {
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
			Owner: ADDR_04_LIKE,
		},
	}
	nftClasses := []NftClass{
		{Id: "likenft1aaaaaa", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/aaaaaa"}},
		{Id: "likenft1bbbbbb", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/bbbbbb"}},
	}
	nfts := []Nft{
		{
			ClassId: nftClasses[0].Id,
			NftId:   "testing-nft-12093810",
			Owner:   ADDR_02_LIKE,
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
	err := InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name   string
		query  QueryCollectorRequest
		owners []string
		counts []int
	}{
		// HACK: default IncludeOwner is true when binding to form
		{
			name:   "empty query",
			query:  QueryCollectorRequest{IncludeOwner: true},
			owners: []string{nfts[1].Owner, nfts[0].Owner},
			counts: []int{2, 1},
		},
		{
			name:   "query with IncludeOwner = false",
			query:  QueryCollectorRequest{IncludeOwner: false},
			owners: []string{nfts[1].Owner},
			counts: []int{2},
		},
		{
			name:   "query by creator, AllIscnVersions = false",
			query:  QueryCollectorRequest{Creator: ADDR_01_COSMOS, IncludeOwner: true},
			owners: []string{},
			counts: []int{},
		},
		{
			name:   "query by creator, AllIscnVersions = true",
			query:  QueryCollectorRequest{Creator: ADDR_01_COSMOS, IncludeOwner: true, AllIscnVersions: true},
			owners: []string{nfts[0].Owner, nfts[1].Owner},
			counts: []int{1, 1},
		},
		{
			name: "query with ignore list",
			query: QueryCollectorRequest{
				IgnoreList:   []string{ADDR_03_COSMOS},
				IncludeOwner: true,
			},
			owners: []string{nfts[0].Owner},
			counts: []int{1},
		},
	}

	p := PageRequest{
		Limit: 10,
	}
NEXT_TESTCASE:
	for i, testCase := range testCases {
		res, err := GetCollector(Conn, testCase.query, p)
		if err != nil {
			t.Errorf("test case #%02d (%s): GetCollector returned error: %#v", i, testCase.name, err)
			continue NEXT_TESTCASE
		}
		if len(res.Collectors) != len(testCase.owners) {
			t.Errorf("test case #%02d (%s): expect len(res.Collectors) = %d, got %d. results = %#v", i, testCase.name, len(testCase.owners), len(res.Collectors), res.Collectors)
			continue NEXT_TESTCASE
		}
		if len(res.Collectors) > 1 {
			prev := res.Collectors[0].Count
			for _, collector := range res.Collectors {
				if collector.Count > prev {
					t.Errorf("test case #%02d (%s): expect Collectors in descending order, got results = %#v", i, testCase.name, res.Collectors)
					continue NEXT_TESTCASE
				}
			}
		}
	NEXT_OWNER:
		for j, owner := range testCase.owners {
			for _, collector := range res.Collectors {
				if collector.Account == owner {
					if collector.Count != testCase.counts[j] {
						t.Errorf("test case #%02d (%s), collector %s: expect count = %d, got %d. results = %#v", i, testCase.name, owner, testCase.counts[j], collector.Count, res.Collectors)
						continue NEXT_TESTCASE
					}
					continue NEXT_OWNER
				}
			}
			t.Errorf("test case #%02d (%s): collector %s not found. results = %#v", i, testCase.name, owner, res.Collectors)
		}
	}
}

func TestCreators(t *testing.T) {
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
			Owner: ADDR_04_LIKE,
		},
	}
	nftClasses := []NftClass{
		{Id: "likenft1aaaaaa", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/aaaaaa"}},
		{Id: "likenft1bbbbbb", Parent: NftClassParent{IscnIdPrefix: "iscn://testing/bbbbbb"}},
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
	err := InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name   string
		query  QueryCreatorRequest
		owners []string
		counts []int
	}{
		// HACK: default IncludeOwner is true when binding to form
		{
			name:   "query by collector (0), AllIscnVersions = false",
			query:  QueryCreatorRequest{Collector: nfts[0].Owner, IncludeOwner: true},
			owners: []string{iscns[1].Owner},
			counts: []int{1},
		},
		{
			name:   "query by collector (0), AllIscnVersions = false, IncludeOwner = false",
			query:  QueryCreatorRequest{Collector: nfts[0].Owner, IncludeOwner: false},
			owners: []string{},
			counts: []int{},
		},
		{
			name:   "query by collector (1), AllIscnVersions = false",
			query:  QueryCreatorRequest{Collector: nfts[1].Owner, IncludeOwner: true},
			owners: []string{iscns[1].Owner, iscns[2].Owner},
			counts: []int{1, 1},
		},
		{
			name:   "query by collector (1), AllIscnVersions = true",
			query:  QueryCreatorRequest{Collector: nfts[1].Owner, IncludeOwner: true, AllIscnVersions: true},
			owners: []string{iscns[0].Owner, iscns[1].Owner, iscns[2].Owner},
			counts: []int{1, 1, 1},
		},
	}

	p := PageRequest{
		Limit: 10,
	}
NEXT_TESTCASE:
	for i, testCase := range testCases {
		res, err := GetCreators(Conn, testCase.query, p)
		if err != nil {
			t.Errorf("test case #%02d (%s): GetCreators returned error: %#v", i, testCase.name, err)
			continue NEXT_TESTCASE
		}
		if len(res.Creators) != len(testCase.owners) {
			t.Errorf("test case #%02d (%s): expect len(res.Creators) = %d, got %d. results = %#v", i, testCase.name, len(testCase.owners), len(res.Creators), res.Creators)
			continue NEXT_TESTCASE
		}
		if len(res.Creators) > 1 {
			prev := res.Creators[0].Count
			for _, creators := range res.Creators {
				if creators.Count > prev {
					t.Errorf("test case #%02d (%s): expect Creators in descending order, got results = %#v", i, testCase.name, res.Creators)
					continue NEXT_TESTCASE
				}
			}
		}
	NEXT_OWNER:
		for j, owner := range testCase.owners {
			for _, creator := range res.Creators {
				if creator.Account == owner {
					if creator.Count != testCase.counts[j] {
						t.Errorf("test case #%02d (%s), creator %s: expect count = %d, got %d. results = %#v", i, testCase.name, owner, testCase.counts[j], creator.Count, res.Creators)
						continue NEXT_TESTCASE
					}
					continue NEXT_OWNER
				}
			}
			t.Errorf("test case #%02d (%s): creator %s not found. results = %#v", i, testCase.name, owner, res.Creators)
		}
	}
}

func TestGetCollectorTopRankedCreators(t *testing.T) {
	defer func() { _ = CleanupTestData(Conn) }()
	r := rand.New(rand.NewSource(19823981948123019))
	iscns := []IscnInsert{}
	nftClasses := []NftClass{}
	nfts := []Nft{}
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
					nfts = append(nfts, Nft{
						ClassId: nftClasses[0].Id,
						NftId:   fmt.Sprintf("%s-%02d", classId, nftIndex),
						Owner:   ADDRS_LIKE[r.Intn(10)],
					})
				}
			}
		}
	}
	err := InsertTestData(DBTestData{Iscns: iscns, NftClasses: nftClasses, Nfts: nfts})
	require.NoError(t, err)

	p := PageRequest{
		Limit:   100,
		Reverse: true,
	}

	shouldBeRankK := func(t *testing.T, res QueryCollectorResponse, collector string, k uint) {
		collectorCount := 0
		for _, collectorEntry := range res.Collectors {
			if collectorEntry.Account == collector {
				collectorCount = collectorEntry.Count
				break
			}
		}
		require.NotZero(t, collectorCount)

		inFrontOfCollectorCount := uint(0)
		for _, collectorEntry := range res.Collectors {
			if collectorEntry.Count > collectorCount {
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
