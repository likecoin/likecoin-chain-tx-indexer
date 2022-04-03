package rest

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestISCNById(t *testing.T) {
	id := "iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1"
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("%s/id?iscn_id=%s", ISCN_ENDPOINT, id),
		nil,
	)
	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}

	t.Log(body)
}

func TestISCNNotExists(t *testing.T) {
	id := "iscn://likecoin-chain/not-exists/1"
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("%s/id?iscn_id=%s", ISCN_ENDPOINT, id),
		nil,
	)
	res, body := request(req)
	if res.StatusCode != 404 {
		t.Fatal("Should return 404", body)
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
