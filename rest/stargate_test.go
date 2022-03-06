package rest

import (
	"net/http/httptest"
	"testing"
)

func TestStargate(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"/cosmos/tx/v1beta1/txs?pagination.limit=3&events=message.module='iscn'", nil)
	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}
}
