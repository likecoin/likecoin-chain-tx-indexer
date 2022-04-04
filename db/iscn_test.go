package db

import (
	"log"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func TestISCNCombineQuery(t *testing.T) {
	tables := []struct {
		query  ISCNRecordQuery
		events types.StringEvents
		length int
	}{
		{
			query: ISCNRecordQuery{
				ContentMetadata: &ContentMetadata{
					Keywords: "Cyberspace,EFF",
				},
			},
			length: 1,
		},
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
		records, err := QueryISCN(conn, v.events, v.query, p)
		if err != nil {
			t.Error(err)
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

func TestQueryISCNByID(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	events := types.StringEvents{
		types.StringEvent{
			Type: "iscn_record",
			Attributes: []types.Attribute{
				{
					Key:   "iscn_id",
					Value: "iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1",
				},
			},
		},
	}
	records, err := QueryISCNByEvents(conn, events)
	if err != nil {
		t.Error(err)
	}
	if len(records) != 1 {
		t.Error("records' length should be 1", records)
	}
}

func TestQueryISCNByOwner(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	events := types.StringEvents{
		types.StringEvent{
			Type: "iscn_record",
			Attributes: []types.Attribute{
				{
					Key:   "owner",
					Value: "cosmos1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j",
				},
			},
		},
	}
	records, err := QueryISCNByEvents(conn, events)
	if err != nil {
		t.Error(err)
	}
	if len(records) == 0 {
		t.Error("records' length should not be 0", records)
	}
}

func TestQueryISCNByRecord(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Release()

	records, err := QueryISCNByRecord(conn, `{"contentFingerprints": ["ar://UbZGi5yn61Iu4o16jM3_cxNgBk61ZRmypr5FabgRKEM"]}`)
	if err != nil {
		t.Error(err)
	}

	if len(records) != 1 {
		t.Error("There should be 1 record", records)
	}
	t.Log(records)
}
