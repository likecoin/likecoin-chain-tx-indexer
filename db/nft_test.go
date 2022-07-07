package db

import "testing"

func TestQueryNftClass(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	p := Pagination{
		Limit: 10,
	}

	res, err := GetNftClass(conn, p)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}

func TestQueryNftByIscn(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	p := Pagination{
		Limit: 10,
	}
	t.Log(p)

	GetNftByIscn(conn, "iscn://likecoin-chain/fIaP4-pj5cdfstg-DsE4_QEMNmzm42PS0uGQ-nPuc_Q")
}
