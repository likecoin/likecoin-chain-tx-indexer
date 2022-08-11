package db

import "testing"

func TestNftMintCount(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		panic(err)
	}
	defer conn.Release()

	table := []struct {
		name string
		q    QueryNftMintCountRequest
	}{
		{name: "empty request"},
		{"include_owner", QueryNftMintCountRequest{IncludeOwner: true}},
		{"ignore_list", QueryNftMintCountRequest{IgnoreList: []string{API_WALLET}}},
		{"include_owner+ignore_lsit", QueryNftMintCountRequest{
			IncludeOwner: true, IgnoreList: []string{API_WALLET}},
		},
	}
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			count, err := GetNftMintCount(conn, v.q)
			if err != nil {
				t.Error(err)
			}
			if count == 0 {
				t.Error("count should not be 0")
			}
			t.Log(count)
		})
	}
}

func TestNftTradeStats(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		panic(err)
	}
	defer conn.Release()

	table := []struct {
		name string
		q    QueryNftTradeStatsRequest
	}{
		{"empty request", QueryNftTradeStatsRequest{API_WALLET}},
	}
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			res, err := GetNftTradeStats(conn, v.q)
			if err != nil {
				t.Error(err)
			}
			t.Log(res)
		})
	}
}
