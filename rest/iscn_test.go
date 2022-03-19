package rest

import (
	"fmt"
	"net/http/httptest"
	"testing"

	iscn "github.com/likecoin/likechain/x/iscn/types"
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
	var result iscn.QueryRecordsByIdResponse
	err := result.Unmarshal([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
