package rest

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"go.uber.org/zap/zapcore"
)

var router *gin.Engine

func env(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestMain(m *testing.M) {
	godotenv.Load("../.env")
	logger.SetupLogger(zapcore.DebugLevel, []string{"stdout"}, "console")
	pool, err := db.NewConnPool(
		env("DB_NAME", "postgres"),
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_USER", "postgres"),
		env("DB_PASS", "password"),
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
	router = getRouter(pool)
	m.Run()
}

func request(req *http.Request) (*http.Response, string) {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}
