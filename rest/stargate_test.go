package rest

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

type Response struct {
	Pagination   interface{}
	Txs          []interface{}
	Tx_responses []interface{}
}

func TestStargate(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		STARGATE_ENDPOINT+"?events=iscn_record.iscn_id='iscn://likecoin-chain/dLbKMa8EVO9RF4UmoWKk2ocUq7IsxMcnQL1_Ps5Vg80/1'", nil)
	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}
	var result Response
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Txs) == 0 {
		t.Fatal("No response:", result)
	}
}

func TestQueryWithKeyword(t *testing.T) {
	term := "LikeCoin"
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("%s?q=%s&pagination.limit=1&events=message.module='iscn'", STARGATE_ENDPOINT, term),
		nil,
	)
	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}
	if !strings.Contains(body, term) {
		t.Fatal("Response doesn't contains the search term", body)
	}
}