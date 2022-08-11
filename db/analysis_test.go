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
		{"ignore_list", QueryNftMintCountRequest{
			IgnoreList: []string{"like17m4vwrnhjmd20uu7tst7nv0kap6ee7js69jfrs"}}},
		{"include_owner+ignore_lsit", QueryNftMintCountRequest{
			IncludeOwner: true, IgnoreList: []string{"like17m4vwrnhjmd20uu7tst7nv0kap6ee7js69jfrs"}},
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
