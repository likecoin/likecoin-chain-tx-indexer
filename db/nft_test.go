package db

import (
	"encoding/json"
	"testing"
)

func TestQueryNftByIscn(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	p := PageRequest{
		Limit: 10,
	}

	q := QueryClassRequest{
		IscnIdPrefix: "iscn://likecoin-chain/fIaP4-pj5cdfstg-DsE4_QEMNmzm42PS0uGQ-nPuc_Q",
	}

	res, err := GetClasses(conn, q, p)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}

func TestQueryNftByOwner(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	q := QueryNftRequest{
		Owner: "like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp",
	}
	p := PageRequest{
		Limit: 1,
	}

	res, err := GetNfts(conn, q, p)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestOwnerByClassId(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	q := QueryOwnerRequest{
		ClassId: "likenft1furc4kuuepyts7ahr0wchc4nev52gkjyeg485vcs9f52snnv0t3s4g0wya",
	}

	res, err := GetOwners(conn, q)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestEventsByNftId(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	q := QueryEventsRequest{
		ClassId: "likenft1ltlz9q5c0xu2xtrjudrgm4emfu37du755kytk8swu4s6yjm268msp6mgf8",
		NftId:   "testing-aurora-86",
	}

	p := PageRequest{
		Limit: 1,
	}

	res, err := GetNftEvents(conn, q, p)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestCollectors(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	q := QueryCollectorRequest{
		Creator: "like156gedr03g3ggwktzhygfusax4df46k8dh6w0me",
	}

	p := PageRequest{
		Limit: 1,
	}

	res, err := GetCollector(conn, q, p)
	if err != nil {
		t.Fatal(err)
	}
	output, _ := json.Marshal(&res)
	t.Log(string(output))
}

func TestCreators(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	q := QueryCreatorRequest{
		Collector: "like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp",
	}

	p := PageRequest{
		Limit: 1,
	}

	res, err := GetCreators(conn, q, p)
	if err != nil {
		t.Fatal(err)
	}
	output, _ := json.Marshal(&res)
	t.Log(string(output))
}
