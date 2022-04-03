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
			query:  "limit=15&page=2",
			length: 15,
		},
		{
			query:  "keywords=DAO&limit=5",
			length: 1,
		},
		{
			query:  "keywords=Cyberspace&keywords=EFF&limit=5",
			length: 1,
		},
		{
			query:  "keywords=香港&owner=cosmos1ykkpc0dnetfsya88f5nrdd7p57kplaw8sva6pj&limit=5",
			length: 5,
		},
		{
			query:  "q=香港&limit=15",
			length: 15,
		},
		{
			query:   "stakeholders.entity.id=John+Perry+Barlow",
			length:  1,
			contain: []string{"John Perry Barlow"},
		},
		{
			query:   "stakeholders.entity.name=Apple+Daily&limit=10",
			length:  10,
			contain: []string{"Apple Daily"},
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
		var records ISCNRecordsResponse
		if err := json.Unmarshal([]byte(body), &records); err != nil {
			t.Fatal(err)
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

func TestISCNByOwner(t *testing.T) {
	owner := "cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j"
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("%s/owner?owner=%s", ISCN_ENDPOINT, owner),
		nil,
	)
	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}

	t.Log(body)
}

func TestISCNByFingerprints(t *testing.T) {
	table := []struct {
		fingerprint string
		statusCode  int
	}{
		{"ipfs://QmPiX4izgDNyJJRnd8V5ei5ce58dsxErpNVre5jcMPBARG", 200},
		{"hash://sha256/d2a92fe4b7c5b9654f8aa303bed0b727931ab44c7f29b2750580abca2cb6597d", 200},
		{"hash://not-exists", 404},
	}
	for _, v := range table {
		req := httptest.NewRequest(
			"GET",
			fmt.Sprintf("%s/fingerprint?fingerprint=%s",
				ISCN_ENDPOINT, v.fingerprint),
			nil,
		)
		res, body := request(req)
		if res.StatusCode != v.statusCode {
			t.Fatalf("expect %d, got %d\n%s", v.statusCode, res.StatusCode, body)
		}
	}
}

func TestISCNCombine(t *testing.T) {
	table := []struct {
		query  string
		status int
	}{
		{
			query:  "iscn_id=iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
			status: 200,
		},
		{
			query:  "owner=cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j",
			status: 200,
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
		if res.StatusCode != v.status {
			t.Fatalf("expect %d, got %d\n%s\n%s", v.status, res.StatusCode, v.query, body)
		}

	}
}
