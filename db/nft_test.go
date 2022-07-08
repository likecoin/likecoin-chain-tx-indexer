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

func TestQueryNftByOwner(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	res, err := GetNftByOwner(conn, "like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
