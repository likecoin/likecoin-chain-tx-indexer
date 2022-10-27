package db

import (
	"encoding/json"
	"testing"
	"time"
)

const KIN = "like13f4glvg80zvfrrs7utft5p68pct4mcq7t5atf6"
const API_WALLET = "like17m4vwrnhjmd20uu7tst7nv0kap6ee7js69jfrs"
const COLLECTOR = "like1xpkwcv48jqdxym26m8f5wjqa3e7dytq3zr733e"
const ISCN_ID_PREFIX = "iscn://likecoin-chain/IKI9PueuJiOsYvhN6z9jPJIm3UGMh17BQ3tEwEzslQo"
const CLASS_ID = "likenft1yhsps5l8tmeuy9y7k0rjpx97cl67cjkjnzkycecw5xrvjjp6c5yqz0ttmc"
const NFT_ID = "writing-4504b0ae-92ee-47db-b2d9-89aa769193ec"

func TestQueryNftClass(t *testing.T) {
	table := []struct {
		name      string
		q         QueryClassRequest
		hasResult bool
	}{
		{
			"iscn_id_prefix",
			QueryClassRequest{IscnIdPrefix: ISCN_ID_PREFIX},
			true,
		},
		{
			"iscn_id_not_exists",
			QueryClassRequest{IscnIdPrefix: "iscn://12341626361263/1"},
			false,
		},
	}
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	p := PageRequest{
		Limit: 10,
	}

	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			res, err := GetClasses(conn, v.q, p)
			if err != nil {
				t.Error(err)
			}
			for _, c := range res.Classes {
				if c.Parent.IscnIdPrefix != v.q.IscnIdPrefix {
					t.Errorf("IscnIdPrefix not equal, expect %s, got %#v", v.q.IscnIdPrefix, c)
				}
				if c.Parent.Account != v.q.Account {
					t.Errorf("Account not equal, expect %s, got %#v", v.q.Account, c)
				}
			}
			if (len(res.Classes) > 0) != v.hasResult {
				t.Errorf("Expect result: %t. Got: %#v. Query: %#v", v.hasResult, res.Classes, v.q)
			}
		})
	}

}

func TestQueryNftByOwner(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	q := QueryNftRequest{
		Owner: KIN,
	}
	p := PageRequest{
		Limit: 1,
	}

	res, err := GetNfts(conn, q, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Nfts) == 0 {
		t.Error("Empty response")
	}
}

func TestOwnerByClassId(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	q := QueryOwnerRequest{
		ClassId: CLASS_ID,
	}

	res, err := GetOwners(conn, q)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Owners) == 0 {
		t.Error("Empty response")
	}
}

func TestNftEvents(t *testing.T) {
	table := []struct {
		name      string
		q         QueryEventsRequest
		hasResult bool
	}{
		{
			"class_id",
			QueryEventsRequest{
				ClassId: CLASS_ID,
			},
			true,
		},
		{
			"class_id + nft_id",
			QueryEventsRequest{
				ClassId: CLASS_ID,
				NftId:   NFT_ID,
			},
			true,
		},
	}

	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	p := PageRequest{
		Limit: 10,
	}
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			res, err := GetNftEvents(conn, v.q, p)
			if err != nil {
				t.Fatal(err)
			}
			if (len(res.Events) > 0) != v.hasResult {
				t.Errorf("Expect hasResult: %t, got %#v, query: %#v", v.hasResult, res.Events, v.q)
			}
		})
	}

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
				Creator:    KIN,
				IgnoreList: []string{API_WALLET},
			},
		},
		{
			"collector_with_ignore",
			QueryRankingRequest{
				Collector:  COLLECTOR,
				IgnoreList: []string{API_WALLET},
			},
		},
		{
			"creator",
			QueryRankingRequest{
				Creator: KIN,
			},
		},
		{
			"collector",
			QueryRankingRequest{
				Collector: COLLECTOR,
			},
		},
		{
			"stakeholder_id",
			QueryRankingRequest{
				StakeholderId: KIN,
			},
		},
		{
			"stakeholder_name",
			QueryRankingRequest{
				StakeholderName: "kin",
			},
		},
		{
			"stakeholder_id_ignore",
			QueryRankingRequest{
				StakeholderId: KIN,
				IgnoreList:    []string{API_WALLET},
			},
		},
		{
			"stakeholder_name_ignore",
			QueryRankingRequest{
				StakeholderName: "kin",
				IgnoreList:      []string{API_WALLET},
			},
		},
		{
			"creation_date",
			QueryRankingRequest{
				After:  time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
				Before: time.Date(2022, 8, 15, 0, 0, 0, 0, time.UTC).Unix(),
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
			// q.IgnoreList = []string{"like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp"}
			res, err := GetClassesRanking(conn, q.QueryRankingRequest, p)
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
		Creator: KIN,
	}
	p := PageRequest{
		Offset:  0,
		Limit:   5,
		Reverse: true,
	}

	res, err := GetCollector(conn, q, p)
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
		Collector: COLLECTOR,
	}
	p := PageRequest{
		Offset: 0,
		Limit:  5,
	}

	res, err := GetCreators(conn, q, p)
	if err != nil {
		t.Fatal(err)
	}
	output, _ := json.MarshalIndent(&res, "", "  ")
	t.Log(string(output))
}
