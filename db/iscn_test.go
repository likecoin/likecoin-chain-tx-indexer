package db

import (
	"testing"
)

func TestISCNCombineQuery(t *testing.T) {
	tables := []struct {
		ISCN
		length int
	}{
		{
			ISCN: ISCN{
				Iscn: "iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
			},
			length: 1,
		},
		{
			ISCN: ISCN{
				Stakeholders: []byte(`[{"name": "kin"}]`),
			},
			length: 5,
		},
		{
			ISCN: ISCN{
				Keywords: Keywords{"DAO"},
			},
		},
		{
			ISCN: ISCN{
				Keywords: Keywords{"Cyberspace", "EFF"},
			},
		},
		{
			ISCN: ISCN{
				Stakeholders: []byte(`[{"name": "John Perry Barlow"}]`),
			},
			length: 1,
		},
		{
			ISCN: ISCN{
				Stakeholders: []byte(`[{"name": "Apple Daily"}]`),
			},
			length: 5,
		},
		{
			ISCN: ISCN{
				Stakeholders: []byte(`[{"name": "《明報》"}]`),
			},
			length: 5,
		},
		{
			ISCN: ISCN{
				Stakeholders: []byte(`[{"name": "depub.SPACE"}]`),
			},
			length: 5,
		},
	}

	for i, v := range tables {
		p := Pagination{
			Limit: 5,
			Page:  1,
			Order: ORDER_DESC,
		}

		records, err := QueryISCN(pool, v.ISCN, p)
		if err != nil {
			t.Fatal(err)
		}
		if v.length != 0 && v.length != len(records) {
			t.Fatalf("%d: There should be %d records, got %d.\n", i, v.length, len(records))
		}
	}
}

func TestISCNList(t *testing.T) {
	p := Pagination{
		Limit: 10,
		Order: ORDER_DESC,
		Page:  1,
	}

	records, err := QueryISCNList(pool, p)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(records))
}

func TestISCNQueryAll(t *testing.T) {
	tables := []struct {
		term   string
		length int
	}{
		{
			term: "",
		},
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
			term: "kin ko",
		},
		{
			term: "LikeCoin",
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
			term: "cosmos1cp3fcmk5ny2c22s0mxut2xefwrdur2t0clgna0",
		},
		{
			term:   "《明報》",
			length: 5,
		},
		{
			term:   "depub.SPACE",
			length: 5,
		},
		{
			term:   "Apple Daily",
			length: 5,
		},
	}

	for _, v := range tables {
		p := Pagination{
			Limit: 5,
			Page:  1,
			Order: ORDER_DESC,
		}

		t.Log(v.term)
		records, err := QueryISCNAll(pool, v.term, p)
		if err != nil {
			t.Fatal(err)
		}
		if v.length != 0 && v.length != len(records) {
			t.Errorf("There should be %d records, got %d.", v.length, len(records))
		}
	}
}
