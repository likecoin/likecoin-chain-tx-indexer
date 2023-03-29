package extractor_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/extractor"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestIscnVersion(t *testing.T) {
	defer CleanupTestData(Conn)
	table := []struct {
		iscn    string
		version int
	}{
		{
			iscn:    "iscn://likecoin-chain/Nj8mKU_TnRFp5kytMF7hJk4_unujhqM0V_9gFrleAgs/1",
			version: 1,
		},
		{
			iscn:    "iscn://likecoin-chain/vxhbRBaMGSdpgaYp7gk7y8iTDMlc6QVZ6XzxaLKGa0U/3",
			version: 3,
		},
		{
			iscn:    "iscn://likecoin-chain/vxhbRBaMGSdpgaYp7gk7y8iTDMlc6QVZ6XzxaLKGa0U",
			version: 0,
		},
	}
	for _, v := range table {
		iscnVersion := extractor.GetIscnVersion(v.iscn)
		require.Equal(t, v.version, iscnVersion)
	}
}

func TestIscn(t *testing.T) {
	defer CleanupTestData(Conn)
	iscnIdPrefix := "iscn://testing/ISCNAAAAAA"
	iscnId1 := iscnIdPrefix + "/1"
	ipld := "ipldxxxxxxxxxx"
	timestamp := time.Unix(123456789, 0)
	recordNotes := `"record notes"`
	stakeholders := []Stakeholder{
		{
			Entity: Entity{Id: "@Apple", Name: "Apple"},
			Data:   []byte(`{"entity":{"id":"@Apple","name":"Apple"},"contributionType":"http://schema.org/author","rewardProportion":9}`),
		},
		{
			Entity: Entity{Id: "@Boy", Name: "Boy"},
			Data:   []byte(`{"entity":{"id":"@Boy","name":"Boy"},"contributionType":"http://schema.org/publisher","rewardProportion":1}`),
		},
	}
	contentMetadata := `{"a": "b", "c": "d"}`
	fingerprints := []string{`hash://testing/AAAAAAAAAA`, `hash://testing/BBBBBBBBBB`}
	txs := []string{
		fmt.Sprintf(
			`{"height":"1234","txhash":"AAAAAA","tx":{"body":{"messages":[{"@type":"/likechain.iscn.MsgCreateIscnRecord","from":"%[1]s","record":{"recordNotes":%[2]s,"contentFingerprints":["%[3]s","%[4]s"],"stakeholders":[%[5]s,%[6]s],"contentMetadata":%[7]s}}],"memo":"AAAAAA"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"iscn_record","attributes":[{"key":"iscn_id","value":"%[8]s"},{"key":"iscn_id_prefix","value":"%[9]s"},{"key":"owner","value":"%[1]s"},{"key":"ipld","value":"%[10]s"}]},{"type":"message","attributes":[{"key":"action","value":"create_iscn_record"},{"key":"sender","value":"%[1]s"}]}]}],"timestamp":"%[11]s"}`,
			ADDR_01_LIKE, recordNotes, fingerprints[0], fingerprints[1], string(stakeholders[0].Data),
			string(stakeholders[1].Data), contentMetadata, iscnId1, iscnIdPrefix, ipld,
			timestamp.UTC().Format(time.RFC3339),
		),
	}
	InsertTestData(DBTestData{
		Txs: txs,
	})

	finished, err := Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	page := PageRequest{Limit: 10}
	res, err := QueryIscn(Conn, IscnQuery{IscnId: iscnId1}, page)
	require.NoError(t, err)
	require.Len(t, res.Records, 1)

	record := res.Records[0]
	require.Equal(t, ipld, record.Ipld)

	data := record.Data
	require.Equal(t, iscnId1, data.Id)
	require.Equal(t, ADDR_01_LIKE, data.Owner)
	require.Truef(t, timestamp.Equal(data.RecordTimestamp), "Record timestamp: expect %s got %s", timestamp, data.RecordTimestamp)
	require.Equal(t, recordNotes, string(data.RecordNotes))

	resStakeholders := []Stakeholder{}
	err = json.Unmarshal(data.Stakeholders, &resStakeholders)
	require.NoError(t, err)
	require.Len(t, resStakeholders, len(stakeholders))
	for i, v := range stakeholders {
		require.Equal(t, v.Entity, resStakeholders[i].Entity)
	}
	require.Equal(t, string(data.ContentMetadata), contentMetadata)

	resFingerprints := []string{}
	err = json.Unmarshal(data.ContentFingerprints, &resFingerprints)
	require.NoError(t, err)
	require.Len(t, resFingerprints, len(resFingerprints))
	for i, v := range fingerprints {
		require.Equal(t, v, resFingerprints[i])
	}

	iscnId2 := iscnIdPrefix + "/2"
	ipld = "ipldyyyyyyyyyy"
	timestamp = time.Unix(123456790, 0)
	recordNotes = `"record notes 2"`
	stakeholders = []Stakeholder{
		{
			Entity: Entity{Id: "@Alpha", Name: "Alpha"},
			Data:   []byte(`{"entity":{"id":"@Alpha","name":"Alpha"},"contributionType":"http://schema.org/author","rewardProportion":7}`),
		},
		{
			Entity: Entity{Id: "@Beta", Name: "Beta"},
			Data:   []byte(`{"entity":{"id":"@Beta","name":"Beta"},"contributionType":"http://schema.org/publisher","rewardProportion":2}`),
		},
		{
			Entity: Entity{Id: "@Gamma", Name: "Gamma"},
			Data:   []byte(`{"entity":{"id":"@Gamma","name":"Gamma"},"contributionType":"http://schema.org/citation","rewardProportion":1}`),
		},
	}
	contentMetadata = `{"e": "f", "g": "h"}`
	fingerprints = []string{`hash://testing/CCCCCCCCCC`}
	txs = []string{
		fmt.Sprintf(
			`{"height":"1235","txhash":"BBBBBB","tx":{"body":{"messages":[{"@type":"/likechain.iscn.MsgUpdateIscnRecord","from":"%[1]s","iscn_id":"%[2]s","record":{"recordNotes":%[3]s,"contentFingerprints":["%[4]s"],"stakeholders":[%[5]s,%[6]s,%[7]s],"contentMetadata":%[8]s}}],"memo":"BBBBBB"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"iscn_record","attributes":[{"key":"iscn_id","value":"%[9]s"},{"key":"iscn_id_prefix","value":"%[10]s"},{"key":"owner","value":"%[1]s"},{"key":"ipld","value":"%[11]s"}]},{"type":"message","attributes":[{"key":"action","value":"update_iscn_record"},{"key":"sender","value":"%[1]s"}]}]}],"timestamp":"%[12]s"}`,
			ADDR_01_LIKE, iscnId1, recordNotes, fingerprints[0], stakeholders[0].Data,
			stakeholders[1].Data, stakeholders[2].Data, contentMetadata, iscnId2, iscnIdPrefix,
			ipld, timestamp.UTC().Format(time.RFC3339),
		),
	}
	InsertTestData(DBTestData{Txs: txs})

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	res, err = QueryIscn(Conn, IscnQuery{IscnId: iscnId2}, page)
	require.NoError(t, err)
	require.Len(t, res.Records, 1)

	record = res.Records[0]
	require.Equal(t, ipld, record.Ipld)

	data = record.Data
	require.Equal(t, iscnId2, data.Id)
	require.Equal(t, ADDR_01_LIKE, data.Owner)
	require.Truef(t, timestamp.Equal(data.RecordTimestamp), "Record timestamp: expect %s got %s", timestamp, data.RecordTimestamp)
	require.Equal(t, recordNotes, string(data.RecordNotes))

	resStakeholders = []Stakeholder{}
	err = json.Unmarshal(data.Stakeholders, &resStakeholders)
	require.NoError(t, err)
	require.Len(t, resStakeholders, len(stakeholders))
	for i, v := range stakeholders {
		require.Equal(t, v.Entity, resStakeholders[i].Entity)
	}
	require.Equal(t, string(data.ContentMetadata), contentMetadata)

	resFingerprints = []string{}
	err = json.Unmarshal(data.ContentFingerprints, &resFingerprints)
	require.NoError(t, err)
	require.Len(t, resFingerprints, len(resFingerprints))
	for i, v := range fingerprints {
		require.Equal(t, v, resFingerprints[i])
	}

	timestamp2 := time.Unix(123456791, 0)
	txs = []string{
		fmt.Sprintf(
			`{"height":"1236","txhash":"CCCCCC","tx":{"body":{"messages":[{"@type":"/likechain.iscn.MsgChangeIscnRecordOwnership","from":"%[1]s","iscn_id":"%[2]s","new_owner":"%[3]s"}],"memo":"CCCCCC"}},"logs":[{"msg_index":0,"log":"","events":[{"type":"iscn_record","attributes":[{"key":"iscn_id","value":"%[2]s"},{"key":"iscn_id_prefix","value":"%[4]s"},{"key":"owner","value":"%[3]s"}]},{"type":"message","attributes":[{"key":"action","value":"msg_change_iscn_record_ownership"},{"key":"sender","value":"%[1]s"}]}]}],"timestamp":"%[5]s"}`,
			ADDR_01_LIKE, iscnId2, ADDR_02_LIKE, iscnIdPrefix, timestamp2.UTC().Format(time.RFC3339),
		),
	}
	InsertTestData(DBTestData{Txs: txs})
	require.NoError(t, err)

	finished, err = Extract(Conn, extractor.ExtractFunc)
	require.NoError(t, err)
	require.True(t, finished)

	res, err = QueryIscn(Conn, IscnQuery{IscnId: iscnId2}, page)
	require.NoError(t, err)
	require.Len(t, res.Records, 1)

	record = res.Records[0]
	require.Equal(t, ipld, record.Ipld)

	data = record.Data
	require.Equal(t, iscnId2, data.Id)
	require.Equal(t, ADDR_02_LIKE, data.Owner)
	require.Truef(t, timestamp.Equal(data.RecordTimestamp), "Record timestamp: expect %s got %s", timestamp, data.RecordTimestamp)
	require.Equal(t, recordNotes, string(data.RecordNotes))

	resStakeholders = []Stakeholder{}
	err = json.Unmarshal(data.Stakeholders, &resStakeholders)
	require.NoError(t, err)
	require.Len(t, resStakeholders, len(stakeholders))
	for i, v := range stakeholders {
		require.Equal(t, v.Entity, resStakeholders[i].Entity)
	}
	require.Equal(t, string(data.ContentMetadata), contentMetadata)

	resFingerprints = []string{}
	err = json.Unmarshal(data.ContentFingerprints, &resFingerprints)
	require.NoError(t, err)
	require.Len(t, resFingerprints, len(resFingerprints))
	for i, v := range fingerprints {
		require.Equal(t, v, resFingerprints[i])
	}
}
