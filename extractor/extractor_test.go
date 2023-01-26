package extractor_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/extractor"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestExtractMultipleEvents(t *testing.T) {
	defer CleanupTestData(Conn)
	iscnIdPrefixA := "iscn://testing/ISCNAAAAAA"
	iscnIdPrefixB := "iscn://testing/ISCNBBBBBB"
	iscnIdA1 := iscnIdPrefixA + "/1"
	iscnIdB1 := iscnIdPrefixB + "/1"
	ipld := "ipldxxxxxxxxxx"
	timestamp := time.Unix(123456789, 0)
	recordNotes := `"record notes"`
	stakeholdersA := []Stakeholder{
		{
			Entity: Entity{Id: "@Apple", Name: "Apple"},
			Data:   []byte(`{"entity":{"id":"@Apple","name":"Apple"},"contributionType":"http://schema.org/author","rewardProportion":9}`),
		},
	}
	stakeholdersB := []Stakeholder{
		{
			Entity: Entity{Id: "@Boy", Name: "Boy"},
			Data:   []byte(`{"entity":{"id":"@Boy","name":"Boy"},"contributionType":"http://schema.org/publisher","rewardProportion":1}`),
		},
	}
	contentMetadata := `{"a": "b", "c": "d"}`
	fingerprintsA := []string{`hash://testing/AAAAAAAAAA`}
	fingerprintsB := []string{`hash://testing/BBBBBBBBBB`}
	txs := []string{
		fmt.Sprintf(
			`{"height":"1234","txhash":"AAAAAA","tx":{"body":{"messages":[{"@type":"/likechain.iscn.MsgCreateIscnRecord","from":"%[1]s","record":{"recordNotes":%[2]s,"contentFingerprints":["%[3]s"],"stakeholders":[%[4]s],"contentMetadata":%[5]s}}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"iscn_record","attributes":[{"key":"iscn_id","value":"%[6]s"},{"key":"iscn_id_prefix","value":"%[7]s"},{"key":"owner","value":"%[1]s"},{"key":"ipld","value":"%[8]s"}]},{"type":"message","attributes":[{"key":"action","value":"create_iscn_record"},{"key":"sender","value":"%[1]s"}]}]}],"timestamp":"%[9]s"}`,
			ADDR_01_LIKE, recordNotes, fingerprintsA[0], string(stakeholdersA[0].Data), contentMetadata, iscnIdA1, iscnIdPrefixA, ipld,
			timestamp.UTC().Format(time.RFC3339),
		),
		fmt.Sprintf(
			`{"height":"1234","txhash":"BBBBBB","tx":{"body":{"messages":[{"@type":"/likechain.iscn.MsgCreateIscnRecord","from":"%[1]s","record":{"recordNotes":%[2]s,"contentFingerprints":["%[3]s"],"stakeholders":[%[4]s],"contentMetadata":%[5]s}}],"memo":"BBBBBB"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"iscn_record","attributes":[{"key":"iscn_id","value":"%[6]s"},{"key":"iscn_id_prefix","value":"%[7]s"},{"key":"owner","value":"%[1]s"},{"key":"ipld","value":"%[8]s"}]},{"type":"message","attributes":[{"key":"action","value":"create_iscn_record"},{"key":"sender","value":"%[1]s"}]}]}],"timestamp":"%[9]s"}`,
			ADDR_01_LIKE, recordNotes, fingerprintsB[0], string(stakeholdersB[0].Data), contentMetadata, iscnIdB1, iscnIdPrefixB, ipld,
			timestamp.UTC().Format(time.RFC3339),
		),
	}
	err := InsertTestData(DBTestData{
		Txs: txs,
	})
	require.NoError(t, err)

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	page := PageRequest{Limit: 10}
	res, err := QueryIscn(Conn, IscnQuery{}, page)
	require.NoError(t, err)
	require.Len(t, res.Records, 2)

	require.Equal(t, iscnIdA1, res.Records[0].Data.Id)
	require.Equal(t, iscnIdB1, res.Records[1].Data.Id)
}

func TestMain(m *testing.M) {
	SetupDbAndRunTest(m, nil)
}
