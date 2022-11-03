package db_test

import (
	"encoding/json"
	"testing"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestNftCount(t *testing.T) {
	conn, err := AcquireFromPool(Pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	table := []struct {
		name string
		q    QueryNftCountRequest
	}{
		{name: "empty request"},
		{"include_owner", QueryNftCountRequest{IncludeOwner: true}},
		{"ignore_list", QueryNftCountRequest{IgnoreList: []string{API_WALLET}}},
		{"include_owner+ignore_lsit", QueryNftCountRequest{
			IncludeOwner: true, IgnoreList: []string{API_WALLET}},
		},
	}
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			count, err := GetNftCount(conn, v.q)
			if err != nil {
				t.Error(err)
			}
			if count.Count == 0 {
				t.Error("count should not be 0")
			}
			t.Log(count)
		})
	}
}

func TestNftTradeStats(t *testing.T) {
	conn, err := AcquireFromPool(Pool)
	if err != nil {
		t.Fatal(err)
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

func TestNftCreatorCount(t *testing.T) {
	conn, err := AcquireFromPool(Pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	count, err := GetNftCreatorCount(conn)
	if err != nil {
		t.Error(err)
	}
	if count.Count == 0 {
		t.Error("should not be 0")
	}
	t.Log(count)
}

func TestNftOwnerCount(t *testing.T) {
	conn, err := AcquireFromPool(Pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	count, err := GetNftOwnerCount(conn)
	if err != nil {
		t.Error(err)
	}
	if count.Count == 0 {
		t.Error("should not be 0")
	}
	t.Log(count)
}

func TestNftOwnerList(t *testing.T) {
	conn, err := AcquireFromPool(Pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	table := []struct {
		name string
		p    PageRequest
	}{
		{"limit 10", PageRequest{Limit: 10}},
		{"limit 10, offset 20", PageRequest{Limit: 10, Offset: 20}},
		{"limit 100", PageRequest{Limit: 100}},
	}
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			res, err := GetNftOwnerList(conn, v.p)
			if err != nil {
				t.Error(err)
			}
			resJson, _ := json.MarshalIndent(&res, "", "  ")
			t.Log(string(resJson))
		})
	}
}
