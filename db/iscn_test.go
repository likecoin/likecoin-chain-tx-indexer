package db_test

import (
	"testing"
	"time"

	iscntypes "github.com/likecoin/likecoin-chain/v3/x/iscn/types"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestIscnCombineQuery(t *testing.T) {
	iscn := IscnInsert{
		Iscn:  "iscn://testing/abcdef/1",
		Owner: ADDR_01_LIKE,
		Stakeholders: []Stakeholder{
			{
				Entity: Entity{Name: "alice", Id: ADDR_01_LIKE},
				Data:   []byte("{}"),
			},
			{
				Entity: Entity{Name: "bob", Id: ADDR_02_LIKE},
				Data:   []byte("{}"),
			},
		},
		Keywords:     []string{"apple", "boy"},
		Fingerprints: []string{"hash://unknown/asdf", "hash://unknown/qwer"},
	}
	err := PrepareTestData([]IscnInsert{iscn}, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

	tables := []struct {
		name string
		IscnQuery
		hasResult bool
	}{
		{"iscn_id", IscnQuery{
			SearchTerm: iscn.Iscn, IscnId: iscn.Iscn,
		}, true},
		{"iscn_id_prefix", IscnQuery{
			SearchTerm: iscn.IscnPrefix, IscnIdPrefix: iscn.IscnPrefix,
		}, true},
		{"owner_like_prefix", IscnQuery{
			SearchTerm: ADDR_01_LIKE, Owner: ADDR_01_LIKE,
		}, true},
		{"owner_cosmos_prefix", IscnQuery{
			SearchTerm: ADDR_01_COSMOS, Owner: ADDR_01_COSMOS,
		}, true},
		{"stakeholder_name_0", IscnQuery{
			SearchTerm:      iscn.Stakeholders[0].Entity.Name,
			StakeholderName: iscn.Stakeholders[0].Entity.Name,
		}, true},
		{"stakeholder_name_1", IscnQuery{
			SearchTerm:      iscn.Stakeholders[1].Entity.Name,
			StakeholderName: iscn.Stakeholders[1].Entity.Name,
		}, true},
		{"stakeholder_id_0", IscnQuery{
			SearchTerm:    iscn.Stakeholders[0].Entity.Id,
			StakeholderId: iscn.Stakeholders[0].Entity.Id,
		}, true},
		{"stakeholder_id_1", IscnQuery{
			SearchTerm:    iscn.Stakeholders[1].Entity.Id,
			StakeholderId: iscn.Stakeholders[1].Entity.Id,
		}, true},
		{"keyword_0", IscnQuery{
			SearchTerm: iscn.Keywords[0],
			Keywords:   []string{iscn.Keywords[0]},
		}, true},
		{"keyword_1", IscnQuery{
			SearchTerm: iscn.Keywords[1],
			Keywords:   []string{iscn.Keywords[1]},
		}, true},
		{"keyword_0&1", IscnQuery{Keywords: iscn.Keywords}, true},
		{"fingerprint_0", IscnQuery{
			SearchTerm:   iscn.Fingerprints[0],
			Fingerprints: []string{iscn.Fingerprints[0]},
		}, true},
		{"fingerprint_1", IscnQuery{
			SearchTerm:   iscn.Fingerprints[1],
			Fingerprints: []string{iscn.Fingerprints[1]},
		}, true},
		{"fingerprint_0&1", IscnQuery{Fingerprints: iscn.Fingerprints}, true},

		{"non_exist_iscn_id", IscnQuery{
			SearchTerm: "iscn://testing/non_exist/1",
			IscnId:     "iscn://testing/non_exist/1",
		}, false},
		{"non_exist_iscn_version", IscnQuery{
			SearchTerm: "iscn://testing/abcdef/2",
			IscnId:     "iscn://testing/abcdef/2",
		}, false},
		{"non_exist_iscn_id_prefix", IscnQuery{
			SearchTerm:   "iscn://testing/non_exist",
			IscnIdPrefix: "iscn://testing/non_exist",
		}, false},
		{"non_exist_owner", IscnQuery{
			SearchTerm: ADDR_03_LIKE, Owner: ADDR_03_LIKE,
		}, false},
		{"non_exist_stakeholder_name", IscnQuery{
			SearchTerm: "non_exist", StakeholderName: "non_exist",
		}, false},
		{"non_exist_stakeholder_id", IscnQuery{
			SearchTerm: "non_exist", StakeholderId: "non_exist",
		}, false},
		{"non_exist_keyword", IscnQuery{
			SearchTerm: "non_exist",
			Keywords:   []string{"non_exist"},
		}, false},
		{"non_exist_fingerprint", IscnQuery{
			SearchTerm:   "hash://unknown/non_exist",
			Fingerprints: []string{"hash://unknown/non_exist"},
		}, false},
	}

	for i, v := range tables {
		t.Run(v.name, func(t *testing.T) {
			for j := 0; j < 2; j++ {
				v.IscnQuery.AllIscnVersions = (j == 0)
				p := PageRequest{
					Limit:   1,
					Reverse: true,
				}
				res, err := QueryIscn(Conn, v.IscnQuery, p)
				if err != nil {
					t.Fatal(err)
				}
				hasResult := (len(res.Records) > 0)
				if hasResult != v.hasResult {
					t.Fatalf("Test %d (%s, AllIscnVersions = %v): hasResult should be %t, got %d results instead. Results = %#v", i, v.name, v.IscnQuery.AllIscnVersions, v.hasResult, len(res.Records), res.Records)
				}

				if v.SearchTerm != "" {
					res, err := QueryIscnSearch(Conn, v.IscnQuery.SearchTerm, p, v.AllIscnVersions)
					if err != nil {
						t.Fatal(err)
					}
					hasResult := (len(res.Records) > 0)
					if hasResult != v.hasResult {
						t.Fatalf("Test %d (%s on QueryIscnSearch, AllIscnVersions = %v): hasResult should be %t, got %d results instead. Results = %#v\n", i, v.name, v.IscnQuery.AllIscnVersions, v.hasResult, len(res.Records), res.Records)
					}
				}
			}
		})
	}
}

func TestIscnQueryLatestVersion(t *testing.T) {
	iscns := []IscnInsert{
		{
			Iscn:  "iscn://testing/abcdef/1",
			Owner: ADDR_01_LIKE,
		},
		{
			Iscn:  "iscn://testing/abcdef/2",
			Owner: ADDR_02_LIKE,
		},
	}
	err := PrepareTestData(iscns, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

	p := PageRequest{
		Limit: 100,
	}

	query := IscnQuery{
		IscnIdPrefix:    "iscn://testing/abcdef",
		AllIscnVersions: true,
	}
	term := "iscn://testing/abcdef"

	res, err := QueryIscn(Conn, query, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records) != 2 {
		t.Fatalf("QueryIscn with AllIscnVersions: true should return 2 records, got %d records", len(res.Records))
	}

	res, err = QueryIscnSearch(Conn, term, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records) != 2 {
		t.Fatalf("QueryIscnSearch with AllIscnVersions: true should return 2 records, got %d records", len(res.Records))
	}

	query.AllIscnVersions = false
	res, err = QueryIscn(Conn, query, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records) != 1 {
		t.Fatalf("QueryIscn with AllIscnVersions: false should only return latest record, got %d records", len(res.Records))
	}
	iscnIdStr := res.Records[0].Data.Id
	iscnId, err := iscntypes.ParseIscnId(iscnIdStr)
	if err != nil {
		t.Fatal(err)
	}
	if iscnId.Version != 2 {
		t.Fatalf("QueryIscn with AllIscnVersions: false should return record with latest version, expect version = 2, got version = %d", iscnId.Version)
	}

	res, err = QueryIscnSearch(Conn, term, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records) != 1 {
		t.Fatalf("QueryIscnSearch with AllIscnVersions: false should only return latest record, got %d records", len(res.Records))
	}
	iscnIdStr = res.Records[0].Data.Id
	iscnId, err = iscntypes.ParseIscnId(iscnIdStr)
	if err != nil {
		t.Fatal(err)
	}
	if iscnId.Version != 2 {
		t.Fatalf("QueryIscnSearch with AllIscnVersions: false should return record with latest version, expect version = 2, got version = %d", iscnId.Version)
	}

	query = IscnQuery{
		Owner:           iscns[0].Owner,
		AllIscnVersions: true,
	}
	term = iscns[0].Owner
	res, err = QueryIscn(Conn, query, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records) != 1 {
		t.Fatalf("QueryIscn old owner with AllIscnVersions: true should return exactly 1 record, got %d records", len(res.Records))
	}
	iscnIdStr = res.Records[0].Data.Id
	iscnId, err = iscntypes.ParseIscnId(iscnIdStr)
	if err != nil {
		t.Fatal(err)
	}
	if iscnId.Version != 1 {
		t.Fatalf("QueryIscn old owner with AllIscnVersions: true should return record with version 1, got version = %d", iscnId.Version)
	}

	res, err = QueryIscnSearch(Conn, term, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records) != 1 {
		t.Fatalf("QueryIscnSearch old owner with AllIscnVersions: true should return exactly 1 record, got %d records", len(res.Records))
	}
	iscnIdStr = res.Records[0].Data.Id
	iscnId, err = iscntypes.ParseIscnId(iscnIdStr)
	if err != nil {
		t.Fatal(err)
	}
	if iscnId.Version != 1 {
		t.Fatalf("QueryIscnSearch old owner with AllIscnVersions: true should return record with version 1, got version = %d", iscnId.Version)
	}

	query.AllIscnVersions = false
	res, err = QueryIscn(Conn, query, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records) > 0 {
		t.Fatalf("QueryIscn old owner with AllIscnVersions: false should return no record, got %d records", len(res.Records))
	}

	res, err = QueryIscnSearch(Conn, term, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records) > 0 {
		t.Fatalf("QueryIscnSearch old owner with AllIscnVersions: false should return no record, got %d records", len(res.Records))
	}
}

func TestIscnList(t *testing.T) {
	iscns := []IscnInsert{
		{Iscn: "iscn://testing/aaaaaa/1"},
		{Iscn: "iscn://testing/aaaaaa/2"},
		{Iscn: "iscn://testing/bbbbbb/1"},
		{Iscn: "iscn://testing/cccccc/1"},
		{Iscn: "iscn://testing/dddddd/1"},
		{Iscn: "iscn://testing/dddddd/2"},
		{Iscn: "iscn://testing/dddddd/3"},
		{Iscn: "iscn://testing/dddddd/4"},
		{Iscn: "iscn://testing/eeeeee/1"},
		{Iscn: "iscn://testing/ffffff/1"},
	}
	err := PrepareTestData(iscns, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

	p := PageRequest{
		Limit: 5,
	}

	res, err := QueryIscnList(Conn, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if (len(res.Records)) != p.Limit {
		t.Fatalf("QueryIscnList (allIscnVersion = true, limit = 5) should return %d results, got %d. reuslts = %#v", p.Limit, len(res.Records), res.Records)
	}

	p.Limit = 100
	res, err = QueryIscnList(Conn, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if (len(res.Records)) != len(iscns) {
		t.Fatalf("QueryIscnList (allIscnVersion = true) should return %d results, got %d. results = %#v", len(iscns), len(res.Records), res.Records)
	}

	res, err = QueryIscnList(Conn, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if (len(res.Records)) != 6 {
		t.Fatalf("QueryIscnList (allIscnVersion = false) should return %d results, got %d. results = %#v", 6, len(res.Records), res.Records)
	}
}

func TestIscnPagination(t *testing.T) {
	iscns := []IscnInsert{
		{Iscn: "iscn://testing/aaaaaa/1", Timestamp: time.Unix(0, 0)},
		{Iscn: "iscn://testing/bbbbbb/1", Timestamp: time.Unix(1, 0)},
		{Iscn: "iscn://testing/cccccc/1", Timestamp: time.Unix(2, 0)},
		{Iscn: "iscn://testing/dddddd/1", Timestamp: time.Unix(3, 0)},
		{Iscn: "iscn://testing/eeeeee/1", Timestamp: time.Unix(4, 0)},
		{Iscn: "iscn://testing/ffffff/1", Timestamp: time.Unix(5, 0)},
		{Iscn: "iscn://testing/gggggg/1", Timestamp: time.Unix(6, 0)},
		{Iscn: "iscn://testing/hhhhhh/1", Timestamp: time.Unix(7, 0)},
		{Iscn: "iscn://testing/iiiiii/1", Timestamp: time.Unix(8, 0)},
		{Iscn: "iscn://testing/jjjjjj/1", Timestamp: time.Unix(9, 0)},
	}
	err := PrepareTestData(iscns, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

	p := PageRequest{Limit: 1}
	res, err := QueryIscnList(Conn, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Pagination.Count != p.Limit {
		t.Errorf("for pagination %#v, expect count = %d, got %d\n", p, p.Limit, res.Pagination.Count)
	}
	if len(res.Records) != res.Pagination.Count {
		t.Errorf(
			"for pagination %#v, expect len(res.Records) == res.Pagination.Count, got %d and %d\n",
			p, len(res.Records), res.Pagination.Count,
		)
	}
	prevTimestamp := res.Records[0].Data.RecordTimestamp
	p.Key = res.Pagination.NextKey
	p.Limit = 6
	res, err = QueryIscnList(Conn, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Pagination.Count != p.Limit {
		t.Errorf("for pagination %#v, expect count = %d, got %d\n", p, p.Limit, res.Pagination.Count)
	}
	if len(res.Records) != res.Pagination.Count {
		t.Errorf(
			"for pagination %#v, expect len(res.Records) == res.Pagination.Count, got %d and %d\n",
			p, len(res.Records), res.Pagination.Count,
		)
	}
	if int(res.Pagination.NextKey) == 0 {
		t.Errorf("for pagination %#v, expect next_key > 0, got 0", p)
	}
	for i, r := range res.Records {
		timestamp := r.Data.RecordTimestamp
		if timestamp.Before(prevTimestamp) {
			t.Errorf(
				"for pagination %#v, expect records in ascending order, but records[%d] has smaller timestamp (%#v) than before (%#v)\n",
				p, i, timestamp, prevTimestamp,
			)
		}
		prevTimestamp = timestamp
	}

	p = PageRequest{Limit: 1, Reverse: true}
	res, err = QueryIscnList(Conn, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Pagination.Count != p.Limit {
		t.Errorf("for pagination %#v, expect count = %d, got %d\n", p, p.Limit, res.Pagination.Count)
	}
	if len(res.Records) != res.Pagination.Count {
		t.Errorf(
			"for pagination %#v, expect len(res.Records) == res.Pagination.Count, got %d and %d\n",
			p, len(res.Records), res.Pagination.Count,
		)
	}
	prevTimestamp = res.Records[0].Data.RecordTimestamp
	p.Key = res.Pagination.NextKey
	p.Limit = 6
	res, err = QueryIscnList(Conn, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Pagination.Count != p.Limit {
		t.Errorf("for pagination %#v, expect count = %d, got %d\n", p, p.Limit, res.Pagination.Count)
	}
	if len(res.Records) != res.Pagination.Count {
		t.Errorf(
			"for pagination %#v, expect len(res.Records) == res.Pagination.Count, got %d and %d\n",
			p, len(res.Records), res.Pagination.Count,
		)
	}
	if int(res.Pagination.NextKey) == 0 {
		t.Errorf("for pagination %#v, expect next_key > 0, got 0", p)
	}
	for i, r := range res.Records {
		timestamp := r.Data.RecordTimestamp
		if timestamp.After(prevTimestamp) {
			t.Errorf(
				"for pagination %#v, expect records in descending order, but records[%d] has greater timestamp (%#v) than before (%#v)\n",
				p, i, timestamp, prevTimestamp,
			)
		}
		prevTimestamp = timestamp
	}
}
