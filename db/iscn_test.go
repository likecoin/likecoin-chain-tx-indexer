package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iscntypes "github.com/likecoin/likecoin-chain/v3/x/iscn/types"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestIscnCombineQuery(t *testing.T) {
	defer CleanupTestData(Conn)
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
	InsertTestData(DBTestData{Iscns: []IscnInsert{iscn}})

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
				require.NoError(t, err)
				if v.hasResult {
					require.NotEmpty(t, res.Records, "Test %d (%s, AllIscnVersions = %v): should have result", i, v.name, v.IscnQuery.AllIscnVersions)
				} else {
					require.Empty(t, res.Records, "Test %d (%s, AllIscnVersions = %v): should not have result", i, v.name, v.IscnQuery.AllIscnVersions)
				}
				if v.SearchTerm != "" {
					res, err := QueryIscnSearch(Conn, v.IscnQuery.SearchTerm, p, v.AllIscnVersions)
					require.NoError(t, err)
					if v.hasResult {
						require.NotEmpty(t, res.Records, "Test %d (%s on QueryIscnSearch, AllIscnVersions = %v): should have result", i, v.name, v.IscnQuery.AllIscnVersions)
					} else {
						require.Empty(t, res.Records, "Test %d (%s on QueryIscnSearch, AllIscnVersions = %v): should not have result", i, v.name, v.IscnQuery.AllIscnVersions)
					}
				}
			}
		})
	}
}

func TestIscnQueryLatestVersion(t *testing.T) {
	defer CleanupTestData(Conn)
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
	InsertTestData(DBTestData{Iscns: iscns})

	p := PageRequest{
		Limit: 100,
	}

	query := IscnQuery{
		IscnIdPrefix:    "iscn://testing/abcdef",
		AllIscnVersions: true,
	}
	term := "iscn://testing/abcdef"

	res, err := QueryIscn(Conn, query, p)
	require.NoError(t, err)
	require.Len(t, res.Records, 2)

	res, err = QueryIscnSearch(Conn, term, p, true)
	require.NoError(t, err)
	require.Len(t, res.Records, 2)

	query.AllIscnVersions = false
	res, err = QueryIscn(Conn, query, p)
	require.NoError(t, err)
	require.Len(t, res.Records, 1)
	iscnIdStr := res.Records[0].Data.Id
	iscnId, err := iscntypes.ParseIscnId(iscnIdStr)
	require.NoError(t, err)
	require.Equal(t, uint64(2), iscnId.Version)

	res, err = QueryIscnSearch(Conn, term, p, false)
	require.NoError(t, err)
	require.Len(t, res.Records, 1)
	iscnIdStr = res.Records[0].Data.Id
	iscnId, err = iscntypes.ParseIscnId(iscnIdStr)
	require.NoError(t, err)
	require.Equal(t, uint64(2), iscnId.Version)

	query = IscnQuery{
		Owner:           iscns[0].Owner,
		AllIscnVersions: true,
	}
	term = iscns[0].Owner
	res, err = QueryIscn(Conn, query, p)
	require.NoError(t, err)
	require.Len(t, res.Records, 1)
	iscnIdStr = res.Records[0].Data.Id
	iscnId, err = iscntypes.ParseIscnId(iscnIdStr)
	require.NoError(t, err)
	require.Equal(t, uint64(1), iscnId.Version)

	res, err = QueryIscnSearch(Conn, term, p, true)
	require.NoError(t, err)
	require.Len(t, res.Records, 1)
	iscnIdStr = res.Records[0].Data.Id
	iscnId, err = iscntypes.ParseIscnId(iscnIdStr)
	require.NoError(t, err)
	require.Equal(t, uint64(1), iscnId.Version)

	query.AllIscnVersions = false
	res, err = QueryIscn(Conn, query, p)
	require.NoError(t, err)
	require.Empty(t, res.Records)

	res, err = QueryIscnSearch(Conn, term, p, false)
	require.NoError(t, err)
	require.Empty(t, res.Records)
}

func TestIscnList(t *testing.T) {
	defer CleanupTestData(Conn)
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
	InsertTestData(DBTestData{Iscns: iscns})

	p := PageRequest{
		Limit: 5,
	}

	res, err := QueryIscnList(Conn, p, true)
	require.NoError(t, err)
	require.Len(t, res.Records, p.Limit)

	p.Limit = 100
	res, err = QueryIscnList(Conn, p, true)
	require.NoError(t, err)
	require.Len(t, res.Records, len(iscns))

	res, err = QueryIscnList(Conn, p, false)
	require.NoError(t, err)
	require.Len(t, res.Records, 6)
}

func TestIscnPagination(t *testing.T) {
	defer CleanupTestData(Conn)
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
	InsertTestData(DBTestData{Iscns: iscns})

	p := PageRequest{Limit: 1}
	res, err := QueryIscnList(Conn, p, true)
	require.NoError(t, err)
	require.Equal(t, p.Limit, res.Pagination.Count)
	require.Len(t, res.Records, res.Pagination.Count)

	prevTimestamp := res.Records[0].Data.RecordTimestamp
	p.Key = res.Pagination.NextKey
	p.Limit = 6
	res, err = QueryIscnList(Conn, p, true)
	require.NoError(t, err)
	require.Equal(t, p.Limit, res.Pagination.Count)
	require.Len(t, res.Records, res.Pagination.Count)
	require.Greater(t, res.Pagination.NextKey, uint64(0))
	for i, r := range res.Records {
		timestamp := r.Data.RecordTimestamp
		require.False(t, timestamp.Before(prevTimestamp),
			"for pagination %#v, expect records in ascending order, but records[%d] has smaller timestamp (%#v) than before (%#v)\n",
			p, i, timestamp, prevTimestamp,
		)
		prevTimestamp = timestamp
	}

	p = PageRequest{Limit: 1, Reverse: true}
	res, err = QueryIscnList(Conn, p, true)
	require.NoError(t, err)
	require.Equal(t, p.Limit, res.Pagination.Count)
	require.Len(t, res.Records, res.Pagination.Count)

	prevTimestamp = res.Records[0].Data.RecordTimestamp
	p.Key = res.Pagination.NextKey
	p.Limit = 6
	res, err = QueryIscnList(Conn, p, true)
	require.NoError(t, err)
	require.Equal(t, p.Limit, res.Pagination.Count)
	require.Len(t, res.Records, res.Pagination.Count)
	require.Greater(t, res.Pagination.NextKey, uint64(0))
	for i, r := range res.Records {
		timestamp := r.Data.RecordTimestamp
		require.False(t, timestamp.After(prevTimestamp),
			"for pagination %#v, expect records in descending order, but records[%d] has greater timestamp (%#v) than before (%#v)\n",
			p, i, timestamp, prevTimestamp,
		)
		prevTimestamp = timestamp
	}
}
