package db_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	MyTime db.NoTimeZoneTime `json:"my_time"`
}

func TestNoTimeZoneTimeUnmarshalJSON(t *testing.T) {
	timeString := "2023-07-12T12:34:56"
	testJson := fmt.Sprintf(`{"my_time": "%s"}`, timeString)
	var ts TestStruct
	if err := json.Unmarshal([]byte(testJson), &ts); err != nil {
		t.Errorf("Error: %v", err)
	}
	expectedTime, _ := time.Parse(db.NoTimeZoneTimeLayout, timeString)
	require.Equal(t, expectedTime, ts.MyTime.Time)
}
