package extractor_test

import (
	"testing"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/extractor"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestExtract(t *testing.T) {
	for i := 0; i < 2; i++ {
		_, err := db.Extract(Conn, extractor.ExtractFunc)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestMain(m *testing.M) {
	SetupDbAndRunTest(m, nil)
}
