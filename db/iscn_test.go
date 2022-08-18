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
			if (len(res.Records) > 0) != v.hasResult {
				t.Fatalf("%d: There should be %t records, got %d.\n", i, v.hasResult, len(res.Records))
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
		Limit:   10,
		Reverse: true,
	}

	res, err := QueryIscnList(conn, p)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(res.Records))
}

func TestIscnQueryAll(t *testing.T) {
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
			res, err := QueryIscnSearch(conn, v.term, p)
			if err != nil {
				t.Fatal(err)
			}
			if (len(res.Records) > 0) != (v.length > 0) {
				t.Errorf("There should be %d records, got %d.", v.length, len(res.Records))
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
		res, err := QueryIscnList(conn, p)
		if err != nil {
			t.Error(err)
		}
		if res.Pagination.NextKey != a {
			t.Error("pagination", p, "expect", a, "got", res.Pagination.NextKey)
		}
	}
}
