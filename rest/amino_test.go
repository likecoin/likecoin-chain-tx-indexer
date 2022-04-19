package rest

import (
	"net/http/httptest"
	"testing"
)

func TestAmino(t *testing.T) {
	table := []struct {
		query  string
		status int
	}{
		{
			query: "/txs?message.module=iscn",
		},
	}

	for _, v := range table {
		req := httptest.NewRequest("GET", v.query, nil)
		res, body := request(req)
		if (v.status != 0 && res.StatusCode != v.status) || res.StatusCode != 200 {
			t.Fatal(body)
		}
	}
}
