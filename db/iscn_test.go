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
	}

	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Error(err)
	}
	defer conn.Release()

	p := Pagination{
		Limit: 5,
		Page:  1,
		Order: ORDER_DESC,
	}

	for _, v := range tables {
		records, err := QueryISCN(conn, v.events, v.query, v.keywords, p)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		switch v.length {
		case 0:
			if len(records) != 0 {
				t.Error("records should be 0", records)
			}

		case 1:
			if len(records) != 1 {
				t.Error("records should be 1", records)
			}

		case 2:
			if len(records) < 2 {
				t.Error("records should be many", records)
			}
		}
		t.Log(len(records))
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
