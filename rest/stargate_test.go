package rest_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/rest"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

type Response struct {
	Pagination   interface{}
	Txs          []interface{}
	Tx_responses []interface{}
}

func TestStargate(t *testing.T) {
	defer CleanupTestData(Conn)
	tx := `
{
  "height": "1",
  "txhash": "AAAAAA",
  "logs": [
    {
      "msg_index": 0,
      "log": "",
      "events": [
        {
          "type": "iscn_record",
          "attributes": [
            {
              "key": "iscn_id",
              "value": "iscn://testing/AAAAAA/1"
            },
            {
              "key": "iscn_id_prefix",
              "value": "iscn://testing/AAAAAA"
            },
            {
              "key": "owner",
              "value": "like1qyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqewmlu9"
            }
          ]
        },
        {
          "type": "message",
          "attributes": [
            { "key": "action", "value": "create_iscn_record" },
            {
              "key": "sender",
              "value": "like1qyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqewmlu9"
            },
            { "key": "module", "value": "iscn" },
            {
              "key": "sender",
              "value": "like1qyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqewmlu9"
            }
          ]
        }
      ]
    }
  ],
  "tx": {
    "@type": "/cosmos.tx.v1beta1.Tx",
    "body": {
      "messages": [
        {
          "@type": "/likechain.iscn.MsgCreateIscnRecord",
          "from": "like1qyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqewmlu9",
          "record": {
            "recordNotes": "",
            "contentFingerprints": [],
            "stakeholders": [],
            "contentMetadata": {}
          }
        }
      ],
      "memo": "AAAAAA",
      "timeout_height": "0",
      "extension_options": [],
      "non_critical_extension_options": []
    },
    "auth_info": { "fee": {} },
    "signatures": [""]
  },
  "timestamp": "2022-01-01T00:00:00Z",
  "events": []
}
`
	InsertTestData(DBTestData{Txs: []string{tx}})

	req := httptest.NewRequest(
		"GET",
		STARGATE_ENDPOINT+"?events=iscn_record.iscn_id='iscn://testing/AAAAAA/1'", nil)
	res, body := request(req)
	require.Equal(t, 200, res.StatusCode)
	var result Response
	require.NoError(t, json.Unmarshal([]byte(body), &result))
	require.NotEmpty(t, result.Txs)
}
