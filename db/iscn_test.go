package db

import (
	"log"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func TestISCNCombineQuery(t *testing.T) {
	tables := []struct {
		query    ISCNRecordQuery
		events   types.StringEvents
		keywords Keywords
		length   int
	}{
		{
			events: []types.StringEvent{
				{
					Type: "iscn_record",
					Attributes: []types.Attribute{
						{
							Key:   "iscn_id",
							Value: "iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
						},
					},
				},
			},
			length: 1,
		},
		{
			query: ISCNRecordQuery{
				Stakeholders: []Stakeholder{
					{
						Entity: &Entity{
							Name: "kin ko",
						},
					},
				},
			},
			length: 2,
		},
		{
			keywords: Keywords{"DAO"},
			length:   1,
		},
		{
			keywords: Keywords{"Cyberspace", "EFF"},
			length:   1,
		},
		{
			query: ISCNRecordQuery{
				Stakeholders: []Stakeholder{
					{
						Entity: &Entity{
							Id: "John Perry Barlow",
						},
					},
				},
			},
			length: 1,
		},
		{
			query: ISCNRecordQuery{
				Stakeholders: []Stakeholder{
					{
						Entity: &Entity{
							Id: "Apple Daily",
						},
					},
				},
			},
			length: 5,
		},
		{
			query: ISCNRecordQuery{
				Stakeholders: []Stakeholder{
					{
						Entity: &Entity{
							Name: "《明報》",
						},
					},
				},
			},
			length: 5,
		},
		{
			query: ISCNRecordQuery{
				Stakeholders: []Stakeholder{
					{
						Entity: &Entity{
							Name: "depub.SPACE",
						},
					},
				},
			},
			length: 5,
		},
	}

	// conn, err := AcquireFromPool(pool)
	// if err != nil {
	// 	t.Error(err)
	// }
	// defer conn.Release()

	for _, v := range tables {
		p := Pagination{
			Limit: uint64(v.length),
			Page:  1,
			Order: ORDER_DESC,
		}

		records, err := QueryISCN(pool, v.events, v.query, v.keywords, p)
		if err != nil {
			t.Fatal(err)
		}
		if v.length != len(records) {
			t.Errorf("There should be %d records, got %d.", v.length, len(records))
		}
	}
}

func TestISCNList(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	p := Pagination{
		Limit: 10,
		Order: ORDER_DESC,
		Page:  1,
	}

	records, err := QueryISCNList(conn, p)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(records))
}
