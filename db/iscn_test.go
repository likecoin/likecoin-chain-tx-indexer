package db

import (
	"log"
	"testing"
)

func TestIscnCombineQuery(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	tables := []struct {
		name string
		IscnQuery
		hasResult bool
	}{
		{
			"iscn_id",
			IscnQuery{
				IscnId: "iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
			},
			true,
		},
		{
			"iscn_id",
			IscnQuery{
				IscnIdPrefix: "iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY",
			},
			true,
		},
		{
			"stakeholder_name",
			IscnQuery{
				StakeholderName: "kin",
			},
			true,
		},
		{
			"keyword",
			IscnQuery{
				Keywords: []string{"LikeCoin"},
			},
			true,
		},
		{
			"keywords",
			IscnQuery{
				Keywords: []string{"superlike", "civicliker"},
			},
			true,
		},
		{
			"John Perry Barlow",
			IscnQuery{
				StakeholderName: "John Perry Barlow",
			},
			true,
		},
		{
			"Apple Daily",
			IscnQuery{
				StakeholderName: "Apple Daily",
			},
			true,
		},
		{
			"Apple Daily ID",
			IscnQuery{
				StakeholderId: "Apple Daily",
			},
			true,
		},
		{
			"Next Digital Ltd",
			IscnQuery{
				StakeholderId: "Next Digital Ltd",
			},
			true,
		},
		{
			"《明報》",
			IscnQuery{
				StakeholderName: "《明報》",
			},
			true,
		},
		{
			"depub.space",
			IscnQuery{
				StakeholderName: "depub.space",
			},
			true,
		},
		{
			"depub.space id",
			IscnQuery{
				StakeholderId: "https://depub.SPACE",
			},
			true,
		},
	}

	for i, v := range tables {
		t.Run(v.name, func(t *testing.T) {
			p := PageRequest{
				Limit:   1,
				Reverse: true,
			}

			res, err := QueryIscn(conn, v.IscnQuery, p)
			if err != nil {
				t.Fatal(err)
			}
			hasResult := (len(res.Records) > 0)
			if hasResult != v.hasResult {
				t.Fatalf("Test %d (%s): hasResult should be %t, got %d results instead.\n", i, v.name, v.hasResult, len(res.Records))
			}

		})
	}
}

func TestIscnList(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	p := PageRequest{
		Limit:   5,
		Reverse: true,
	}

	res, err := QueryIscnList(conn, p)
	if err != nil {
		t.Fatal(err)
	}
	if (len(res.Records)) != p.Limit {
		t.Fatalf("QueryIscnList should return %d results, got %d.\n", p.Limit, len(res.Records))
	}
}

func TestIscnQueryAll(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	tables := []struct {
		term      string
		hasResult bool
	}{
		{
			term:      "0xNaN",
			hasResult: true,
		},
		{
			term: `" OR "1"="`,
			// SQL injection test
		},
		{
			term: "itdoesnotexists",
		},
		{
			term:      "kin ko",
			hasResult: true,
		},
		{
			term:      "LikeCoin",
			hasResult: true,
		},
		{
			term:      "ar://3sTMJ3K8ZQMuDMcJmfSkJT5xQfBF7U6kUDnnowN3X84",
			hasResult: true,
		},
		{
			term:      "iscn://likecoin-chain/vbLIsrIZVEkFRHEoFJX3LXPszH5oqzrMld32XrbxVgU/1",
			hasResult: true,
		},
		{
			term:      "iscn://likecoin-chain/vbLIsrIZVEkFRHEoFJX3LXPszH5oqzrMld32XrbxVgU",
			hasResult: true,
		},
		{
			term:      "cosmos1ykkpc0dnetfsya88f5nrdd7p57kplaw8sva6pj",
			hasResult: true,
		},
		{
			term:      "《明報》",
			hasResult: true,
		},
		{
			term:      "depub.space",
			hasResult: true,
		},
		{
			term:      "Apple Daily",
			hasResult: true,
		},
	}

	p := PageRequest{
		Limit: 1,
	}
	for i, v := range tables {
		t.Run(v.term, func(t *testing.T) {
			t.Log(v.term)
			res, err := QueryIscnSearch(conn, v.term, p)
			if err != nil {
				t.Fatal(err)
			}
			hasResult := (len(res.Records) > 0)
			if hasResult != v.hasResult {
				t.Fatalf("Test %d (%s): hasResult should be %t, got %d results instead.\n", i, v.term, v.hasResult, len(res.Records))
			}
		})
	}
}

func TestIscnPagination(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	table := map[PageRequest]uint64{
		{
			Key:   2,
			Limit: 3,
		}: 5,
		{
			Key:     3,
			Limit:   2,
			Reverse: true,
		}: 1,
	}
	for p, expectedNextKey := range table {
		res, err := QueryIscnList(conn, p)
		if err != nil {
			t.Fatal(err)
		}
		if res.Pagination.Count != p.Limit {
			t.Errorf("for pagination %##v, expect count = %d, got %d\n", p, p.Limit, res.Pagination.Count)
		}
		if len(res.Records) != res.Pagination.Count {
			t.Errorf(
				"for pagination %##v, expect len(res.Records) == res.Pagination.Count, got %d and %d\n",
				p, len(res.Records), res.Pagination.Count,
			)
		}
		if res.Pagination.NextKey != expectedNextKey {
			t.Errorf("for pagination %##v, expect next key = %d, got %d\n", p, expectedNextKey, res.Pagination.NextKey)
		}
	}
}
