package rest

import (
	"net/http/httptest"
	"testing"
)

func TestAmino(t *testing.T) {
	req := httptest.NewRequest("GET", "/txs?message.module=iscn", nil)
	result := string(request(req))
	t.Log(result)
}
