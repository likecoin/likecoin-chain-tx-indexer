package rest

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestISCNCombine(t *testing.T) {
	table := []struct {
		name    string
		query   string
		status  int
		length  int
		contain []string
	}{
		{
			name:  "iscn_id",
			query: "iscn_id=iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
		},
		{
			name:  "owner",
			query: "owner=cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j",
		},
		{
			name:  "fingerprint",
			query: "fingerprint=hash://sha256/8e6984120afb8ac0f81080cf3e7a38042c472c4deb3e2588480df6e199741c89",
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
			name:   "keyword",
			query:  "keyword=DAO&limit=5",
			length: 2,
		},
		{
			name:  "multiple-keywords",
			query: "keyword=Cyberspace&keyword=EFF&limit=5",
		},
		{
			name:   "keyword & owner",
			query:  "keyword=香港&owner=cosmos1ykkpc0dnetfsya88f5nrdd7p57kplaw8sva6pj&limit=5",
			length: 1,
		},
		{
			name:    "stakeholder name",
			query:   "stakeholder.name=John+Perry+Barlow",
			length:  1,
			contain: []string{"John Perry Barlow"},
		},
		{
			name:    "Apple Daily",
			query:   "stakeholder.name=Apple+Daily&limit=10",
			length:  1,
			contain: []string{"Apple Daily"},
		},
		{
			name:    "search by iscn_id",
			query:   "q=iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
			contain: []string{"iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1"},
		},
		{
			name:   "search by keyword",
			query:  "q=香港&limit=5",
			length: 1,
		},
		{
			name:    "search by owner",
			query:   "q=cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j",
			contain: []string{"cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j"},
		},
		{
			name:    "search by fingerprint",
			query:   "q=hash://sha256/8e6984120afb8ac0f81080cf3e7a38042c472c4deb3e2588480df6e199741c89",
			contain: []string{"hash://sha256/8e6984120afb8ac0f81080cf3e7a38042c472c4deb3e2588480df6e199741c89"},
		},
		{
			name:    "LikeCoin",
			query:   "q=LikeCoin",
			contain: []string{"LikeCoin"},
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
			for _, record := range records.Records {
				if v.contain != nil {
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
