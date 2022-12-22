package rest_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/rest"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestISCNCombine(t *testing.T) {
	iscns := []db.IscnInsert{
		{
			Iscn:  "iscn://testing/abcdef/1",
			Owner: ADDR_01_LIKE,
			Stakeholders: []db.Stakeholder{
				{
					Entity: db.Entity{Name: "alice", Id: ADDR_01_LIKE},
					Data:   []byte("{}"),
				},
				{
					Entity: db.Entity{Name: "bob", Id: ADDR_02_LIKE},
					Data:   []byte("{}"),
				},
			},
			Keywords:     []string{"apple", "boy"},
			Fingerprints: []string{"hash://unknown/asdf", "hash://unknown/qwer"},
		},
	}
	err := InsertTestData(DBTestData{Iscns: iscns})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = CleanupTestData(Conn) }()

	table := []struct {
		name    string
		query   string
		status  int
		length  int
		contain []string
	}{
		{
			name:  "iscn_id",
			query: "iscn_id=" + iscns[0].Iscn,
		},
		{
			name:  "owner",
			query: "owner=" + iscns[0].Owner,
		},
		{
			name:  "fingerprint (0)",
			query: "fingerprint=" + iscns[0].Fingerprints[0],
		},
		{
			name:  "fingerprint (1)",
			query: "fingerprint=" + iscns[0].Fingerprints[1],
		},
		{
			name:  "empty",
			query: "",
		},
		{
			name:  "limit",
			query: "limit=15",
		},
		{
			name:  "keyword (0)",
			query: "keyword=" + iscns[0].Keywords[0],
		},
		{
			name:  "keyword (1)",
			query: "keyword=" + iscns[0].Keywords[1],
		},
		{
			name:  "keyword (0,1)",
			query: "keyword=" + iscns[0].Keywords[0] + "&keyword=" + iscns[0].Keywords[1],
		},
		{
			name:  "keyword & owner",
			query: "keyword=" + iscns[0].Keywords[0] + "&owner=" + iscns[0].Owner,
		},
		{
			name:  "stakeholder name (0)",
			query: "stakeholder.name=" + iscns[0].Stakeholders[0].Entity.Name,
		},
		{
			name:  "stakeholder name (0)",
			query: "stakeholder.name=" + iscns[0].Stakeholders[1].Entity.Name,
		},
		{
			name:  "stakeholder id (0)",
			query: "stakeholder.id=" + iscns[0].Stakeholders[0].Entity.Id,
		},
		{
			name:  "stakeholder id (1)",
			query: "stakeholder.id=" + iscns[0].Stakeholders[1].Entity.Id,
		},
		{
			name:  "search by iscn_id",
			query: "q=" + iscns[0].Iscn,
		},
		{
			name:  "search by keyword",
			query: "q=" + iscns[0].Keywords[0],
		},
		{
			name:  "search by owner",
			query: "q=" + iscns[0].Owner,
		},
		{
			name:  "search by fingerprint",
			query: "q=" + iscns[0].Fingerprints[0],
		},
	}
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			req := httptest.NewRequest(
				"GET",
				fmt.Sprintf("%s?%s",
					ISCN_ENDPOINT, v.query),
				nil,
			)
			res, body := request(req)
			if v.status == 0 {
				v.status = 200
			}
			if res.StatusCode != v.status {
				t.Fatalf("expect %d, got %d\n%s\n%s", v.status, res.StatusCode, v.query, body)
			}
			if res.StatusCode != 200 {
				return
			}
			var records struct {
				Records []json.RawMessage
			}

			if err := json.Unmarshal([]byte(body), &records); err != nil {
				t.Fatal(err)
			}
			if len(records.Records) == 0 {
				t.Fatalf("No response, %s", body)
				return
			}
			if v.length != 0 && len(records.Records) != v.length {
				t.Errorf("Length should be %d, got %d.\n%s\n", v.length, len(records.Records), v.query)
			}
			if v.contain != nil {
				for _, record := range records.Records {
					for _, s := range v.contain {
						if !strings.Contains(string(record), s) {
							t.Errorf("record should contain %s, but not found: %s", s, string(record))
						}
					}
				}

			}
		})
	}
}
