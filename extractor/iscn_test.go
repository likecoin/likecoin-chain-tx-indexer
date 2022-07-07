package extractor

import (
	"log"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	. "github.com/likecoin/likecoin-chain-tx-indexer/utils"
	"go.uber.org/zap/zapcore"
)

var pool *pgxpool.Pool

func TestConvert(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	for i := 0; i < 2; i++ {
		_, err := extractISCN(conn)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestIscnVersion(t *testing.T) {
	table := []struct {
		iscn    string
		version int
	}{
		{
			iscn:    "iscn://likecoin-chain/Nj8mKU_TnRFp5kytMF7hJk4_unujhqM0V_9gFrleAgs/1",
			version: 1,
		},
		{
			iscn:    "iscn://likecoin-chain/vxhbRBaMGSdpgaYp7gk7y8iTDMlc6QVZ6XzxaLKGa0U/3",
			version: 3,
		},
		{
			iscn:    "iscn://likecoin-chain/vxhbRBaMGSdpgaYp7gk7y8iTDMlc6QVZ6XzxaLKGa0U",
			version: 0,
		},
	}
	for _, v := range table {
		if a := getIscnVersion(v.iscn); a != v.version {
			t.Errorf("parse %s expect %d got %d\n", v.iscn, v.version, a)
		}
	}
}

func TestMain(m *testing.M) {
	godotenv.Load("../.env")
	logger.SetupLogger(zapcore.DebugLevel, []string{"stdout"}, "console")
	var err error
	pool, err = NewConnPool(
		Env("DB_NAME", "postgres"),
		Env("DB_HOST", "localhost"),
		Env("DB_PORT", "5432"),
		Env("DB_USER", "postgres"),
		Env("DB_PASS", "password"),
		32,
		4,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer pool.Close()
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}

	err = InitDB(conn)
	if err != nil {
		log.Fatalln(err)
	}
	conn.Release()
	m.Run()
}
