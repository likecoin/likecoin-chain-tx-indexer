package rest

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

var router *gin.Engine

func TestMain(m *testing.M) {
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

	router.GET("/test", func(c *gin.Context) {
		c.String(200, "Yes, it's working")
	})
	router.GET("/txs", func(c *gin.Context) {
		handleAminoTxsSearch(c, pool)
	})
	m.Run()
}

func request(req *http.Request) string {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	return string(body)
}

func TestReq(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	result := request(req)
	log.Println(result)
}
