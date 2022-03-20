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
