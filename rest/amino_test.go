package rest

import (
	"net/http/httptest"
	"testing"
)

func TestAmino(t *testing.T) {
	req := httptest.NewRequest("GET", "/txs?message.module=iscn", nil)
	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}
}
