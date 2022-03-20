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
