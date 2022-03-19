package rest

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"go.uber.org/zap/zapcore"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	logger.SetupLogger(zapcore.DebugLevel, []string{"stdout"}, "console")
	pool, err := db.NewConnPool(
		"mydb",
		"localhost",
		"5432",
		"wancat",
		"password",
		32,
		4,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer pool.Close()
	conn, err := db.AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	db.InitDB(conn)
	router = gin.Default()

	router.GET("/test", func(ctx *gin.Context) {
		ctx.String(200, "Yes, it's working")
	})
	router.GET("/txs", func(ctx *gin.Context) {
		handleAminoTxsSearch(ctx, pool)
	})
	router.GET("/cosmos/tx/v1beta1/txs", func(ctx *gin.Context) {
		handleStargateTxsSearch(ctx, pool)
	})
	m.Run()
}

func request(req *http.Request) (*http.Response, string) {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}

func TestReq(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	res, body := request(req)
	if res.StatusCode != 200 {
		t.Fatal(body)
	}
}