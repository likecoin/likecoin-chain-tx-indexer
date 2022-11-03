package rest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"

	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	SetupDbAndRunTest(m, func(pool *pgxpool.Pool) {
		router = getRouter(pool)
	})
}

func request(req *http.Request) (*http.Response, string) {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}
