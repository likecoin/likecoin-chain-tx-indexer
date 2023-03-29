package rest_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestAmino(t *testing.T) {
	defer CleanupTestData(Conn)
	b := db.NewBatch(Conn, 10000)
	b.Batch.Queue(
		"INSERT INTO txs (height, tx_index, tx, events) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", 1, 1,
		[]byte(`
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
`),
		[]string{`message.module="iscn"`},
	)
	err := b.Flush()
	require.NoError(t, err)

	table := []struct {
		query  string
		status int
	}{
		{
			query: "/txs?message.module=iscn",
		},
	}

	for _, v := range table {
		req := httptest.NewRequest("GET", v.query, nil)
		res, body := request(req)
		require.Equal(t, 200, res.StatusCode)
		if v.status != 0 {
			require.Equal(t, v.status, res.StatusCode, "body: %s", body)
		}
	}
}
