package db_test

import (
	"testing"
	"time"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	SetupDbAndRunTest(m, nil)
}

func TestBlockTime(t *testing.T) {
	timeStr := "2018-01-01T00:00:00Z"
	b := NewBatch(Conn, 10000)
	b.UpdateLatestBlockTime(timeStr)
	b.Flush()
	blockTime, err := GetLatestBlockTime(Conn)
	require.NoError(t, err)
	require.Equal(t, timeStr, blockTime.Format(time.RFC3339))
}
