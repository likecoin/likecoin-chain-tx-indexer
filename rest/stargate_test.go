package rest

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStargate(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"/cosmos/tx/v1beta1/txs?pagination.limit=3&events=iscn_record.iscn_id='iscn://likecoin-chain/dLbKMa8EVO9RF4UmoWKk2ocUq7IsxMcnQL1_Ps5Vg80/1'", nil)
	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}
	t.Log(body)
}

func TestQueryWithTerm(t *testing.T) {
	term := "LikeCoin"
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("/cosmos/tx/v1beta1/txs?q=%s&pagination.limit=1&events=message.module='iscn'", term),
		nil,
	)
	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}
	if !strings.Contains(body, term) {
		t.Fatal("Response doesn't contains the search term", body)
	}
	t.Log(body)
}
