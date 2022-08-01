package db

import (
	"encoding/json"
	"testing"
	"time"
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

func TestQueryNftRanking(t *testing.T) {
	table := []struct {
		name string
		QueryRankingRequest
	}{
		{
			"type",
			QueryRankingRequest{
				Type: "CreativeWork",
			},
		},
		{
			"creator_with_ignore",
			QueryRankingRequest{
				Creator:    "like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf",
				IgnoreList: []string{"like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp"},
			},
		},
		{
			"collector_with_ignore",
			QueryRankingRequest{
				Collector:  "like13v8qtt0jz6y2304559v7l29sy7prz50jqwdewn",
				IgnoreList: []string{"like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp"},
			},
		},
		{
			"creator",
			QueryRankingRequest{
				Creator: "like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf",
			},
		},
		{
			"collector",
			QueryRankingRequest{
				Collector: "like13v8qtt0jz6y2304559v7l29sy7prz50jqwdewn",
			},
		},
		{
			"stakeholder_id",
			QueryRankingRequest{
				StakeholderId: "did:like:1shkl5gqzxcs9yh3qjdeggaz3yg5s83754dx2dh",
			},
		},
		{
			"stakeholder_name",
			QueryRankingRequest{
				StakeholderName: "Author",
			},
		},
		{
			"creation_date",
			QueryRankingRequest{
				After:  time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC).Unix(),
				Before: time.Date(2022, 7, 15, 0, 0, 0, 0, time.UTC).Unix(),
			},
		},
	}
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	p := PageRequest{
		Limit: 100,
	}

	for _, q := range table {
		t.Run(q.name, func(t *testing.T) {
			q.PageRequest = p
			// q.IgnoreList = []string{"like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp"}
			res, err := GetClassesRanking(conn, q.QueryRankingRequest)
			if err != nil {
				t.Error(err)
			}
			input, _ := json.MarshalIndent(&q, "", "  ")
			output, _ := json.MarshalIndent(&res, "", "  ")
			if len(res.Classes) == 0 {
				t.Error("No response", string(input), string(output))
				return
			}
			// t.Log(string(input), string(output))
		})
	}
}
func TestCollectors(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	q := QueryCollectorRequest{
		Creator: "like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf",
		PageRequest: PageRequest{
			Offset:  0,
			Limit:   5,
			Reverse: true,
		},
	}

	res, err := GetCollector(conn, q)
	if err != nil {
		t.Fatal(err)
	}
	output, _ := json.MarshalIndent(&res, "", "  ")
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
		PageRequest: PageRequest{
			Offset: 0,
			Limit:  5,
		},
	}

	res, err := GetCreators(conn, q)
	if err != nil {
		t.Fatal(err)
	}
	output, _ := json.MarshalIndent(&res, "", "  ")
	t.Log(string(output))
}
