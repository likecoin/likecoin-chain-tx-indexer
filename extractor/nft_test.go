package extractor

import (
	"testing"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func TestNFT(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	GetNFTTxs(conn)
}
