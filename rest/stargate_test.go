package rest

import (
	"net/http/httptest"
	"testing"
)

func TestStargate(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"/cosmos/tx/v1beta1/txs?pagination.limit=3&events=message.module='iscn'", nil)
	result := string(request(req))
	t.Log(result)
}
