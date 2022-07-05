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
		query   string
		status  int
		length  int
		contain []string
	}{
		{
			query: "iscn_id=iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
		},
		{
			query: "owner=cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j",
		},
		{
			query: "fingerprint=hash://sha256/8e6984120afb8ac0f81080cf3e7a38042c472c4deb3e2588480df6e199741c89",
		},
		{
			query: "",
		},
		{
			query: "limit=15&page=2",
		},
		{
			query:  "keywords=DAO&limit=5",
			length: 5,
		},
		{
			query: "keywords=Cyberspace&keywords=EFF&limit=5",
		},
		{
			query:  "keywords=香港&owner=cosmos1ykkpc0dnetfsya88f5nrdd7p57kplaw8sva6pj&limit=5",
			length: 5,
		},
		{
			query: "q=香港&limit=15",
		},
		{
			query:   "stakeholders.entity.name=John+Perry+Barlow",
			length:  1,
			contain: []string{"John Perry Barlow"},
		},
		{
			query:   "stakeholders.entity.name=Apple+Daily&limit=10",
			length:  10,
			contain: []string{"Apple Daily"},
		},
		{
			query:   "q=iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
			contain: []string{"iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1"},
		},
		{
			query:   "q=cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j",
			contain: []string{"cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j"},
		},
		{
			query:   "q=hash://sha256/8e6984120afb8ac0f81080cf3e7a38042c472c4deb3e2588480df6e199741c89",
			contain: []string{"hash://sha256/8e6984120afb8ac0f81080cf3e7a38042c472c4deb3e2588480df6e199741c89"},
		},
		{
			query:   "q=LikeCoin",
			contain: []string{"LikeCoin"},
		},
	}
	for _, v := range table {
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
			continue
		}
		var records ISCNRecordsResponse
		if err := json.Unmarshal([]byte(body), &records); err != nil {
			t.Fatal(err)
		}
		if len(records.Records) == 0 {
			t.Errorf("No response, %s", body)
			continue
		}
		if v.length != 0 && len(records.Records) != v.length {
			t.Errorf("Length should be %d, got %d.\n%s\n", v.length, len(records.Records), v.query)
		}
		for _, record := range records.Records {
			if v.contain != nil {
				for _, s := range v.contain {
					if !strings.Contains(record.String(), s) {
						t.Errorf("record should contain %s, but not found: %s", s, record.String())
					}
				}
			}

		}
	}
}
