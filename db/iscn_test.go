package db

import (
	"log"
	"testing"
)

func TestISCNCombineQuery(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	tables := []struct {
		name string
		ISCNQuery
		hasResult bool
	}{
		{
			"iscn_id",
			ISCNQuery{
				IscnID: "iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
			},
			true,
		},
		{
			"stakeholder_name",
			ISCNQuery{
				StakeholderName: "kin",
			},
			true,
		},
		{
			"keyword",
			ISCNQuery{
				Keywords: []string{"LikeCoin"},
			},
			true,
		},
		{
			"keywords",
			ISCNQuery{
				Keywords: []string{"superlike", "civicliker"},
			},
			true,
		},
		{
			"John Perry Barlow",
			ISCNQuery{
				StakeholderName: "John Perry Barlow",
			},
			true,
		},
		{
			"Apple Daily",
			ISCNQuery{
				StakeholderName: "Apple Daily",
			},
			true,
		},
		{
			"Apple Daily ID",
			ISCNQuery{
				StakeholderID: "Apple Daily",
			},
			true,
		},
		{
			"Next Digital Ltd",
			ISCNQuery{
				StakeholderID: "Next Digital Ltd",
			},
			true,
		},
		{
			"《明報》",
			ISCNQuery{
				StakeholderName: "《明報》",
			},
			true,
		},
		{
			"depub.space",
			ISCNQuery{
				StakeholderName: "depub.space",
			},
			true,
		},
		{
			"depub.space id",
			ISCNQuery{
				StakeholderID: "https://depub.SPACE",
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

			res, err := QueryISCN(conn, v.ISCNQuery, p)
			if err != nil {
				t.Fatal(err)
			}
			if (len(res.Records) > 0) != v.hasResult {
				t.Fatalf("%d: There should be %t records, got %d.\n", i, v.hasResult, len(res.Records))
			}

		})
	}
}

func TestISCNList(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	p := PageRequest{
		Limit:   10,
		Reverse: true,
	}

	res, err := QueryISCNList(conn, p)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(res.Records))
}

func TestISCNQueryAll(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	tables := []struct {
		term   string
		length int
	}{
		{
			term:   "0xNaN",
			length: 5,
		},
		{
			term: `" OR "1"="`,
			// SQL injection test
		},
		{
			term: "itdoesnotexists",
		},
		{
			term:   "kin ko",
			length: 5,
		},
		{
			term:   "LikeCoin",
			length: 5,
		},
		{
			term:   "ar://3sTMJ3K8ZQMuDMcJmfSkJT5xQfBF7U6kUDnnowN3X84",
			length: 1,
		},
		{
			term:   "iscn://likecoin-chain/zGY8c7obhwx7qa4Ro763kr6lvBCZ4SIMagYRXRXYSnM/1",
			length: 1,
		},
		{
			term:   "cosmos1cp3fcmk5ny2c22s0mxut2xefwrdur2t0clgna0",
			length: 5,
		},
		{
			term:   "《明報》",
			length: 5,
		},
		{
			term:   "depub.space",
			length: 5,
		},
		{
			term:   "Apple Daily",
			length: 100,
		},
	}

	p := PageRequest{
		Limit: 1,
	}
	for _, v := range tables {
		t.Run(v.term, func(t *testing.T) {
			t.Log(v.term)
			res, err := QueryISCNAll(conn, v.term, p)
			if err != nil {
				t.Fatal(err)
			}
			if (len(res.Records) > 0) != (v.length > 0) {
				t.Errorf("There should be %d records, got %d.", v.length, len(res.Records))
			}
		})
	}
}

func TestISCNPagination(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	table := map[PageRequest]uint64{
		{
			Key:   1300,
			Limit: 10,
		}: 1310,
		{
			Key:     1300,
			Limit:   10,
			Reverse: true,
		}: 1290,
	}
	for p, a := range table {
		res, err := QueryISCNList(conn, p)
		if err != nil {
			t.Error(err)
		}
		if res.Pagination.NextKey != a {
			t.Error("pagination", p, "expect", a, "got", res.Pagination.NextKey)
		}
	}
}
