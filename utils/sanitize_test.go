package utils

import (
	"bytes"
	"strings"
	"testing"
)

func TestSanitizeJSON(t *testing.T) {
	input := []byte(`"\t\\\b\c\d\e\r\n\uD83D\uDe02\"\/\f\b\uDe02\uD83D\u0000\u0001XXXX\u012X"`)
	output := SanitizeJSON(input)
	expected := []byte(`"\t\\\bcde\r\n\uD83D\uDe02\"\/\f\b\u0001XXXX012X"`)
	if !bytes.Equal(output, expected) {
		t.Errorf("sanitizeJSON failed, expect %#v, got %#v", string(expected), string(output))
	}

	input2 := `
{
  "tx": {
    "body": {
      "messages": [
        {
          "@type": "/likechain.iscn.MsgCreateIscnRecord",
          "from": "like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn",
          "record": {
            "recordNotes": "like13f30lx5q8dscgxpdfpqkakx8ejwgstxqgda0s6",
            "contentFingerprints": [
              "hash://sha256/bafyreia4rcfbi3heobobu7sysnyfvtwplt5wgooxe2spqexkhijs4juruq",
              "ipfs://QmRrVSxdrCRSQ5jb5A7BCY8aReFMqhtUzuaicnE5C7vZbP"
            ],
            "stakeholders": [
              {"entity":{"@id":"did:cosmos:cosmos13f30lx5q8dscgxpdfpqkakx8ejwgstxqm3pdnp","name":"困天索 (@jerryokla198)"},"rewardProportion":100,"contributionType":"http://schema.org/author"},
              {"rewardProportion":0,"contributionType":"http://schema.org/publisher","entity":{"name":"Matters.News"}}
            ],
            "contentMetadata": {"@context":"http://schema.org/","@type":"Article","name":"分享一個可以做且會賺錢的生意","description":"大家好！跟大家分享一個近3年來很罕見 我認為可以做且會賺錢的生意！這是結合酒店、經紀傳播、博弈及超跑租賃業，所整合的一個「創新電子化高端酒店」的生意  目前有許多同業都非常認同這個案子  我認為 它不只是 台灣第一  甚至是     全球第一  如果有錢希望大家可以了解看看  \ud83d...","version":1,"url":"https://matters.news/@jerryokla198/328574-分享一個可以做且會賺錢的生意-bafyreia4rcfbi3heobobu7sysnyfvtwplt5wgooxe2spqexkhijs4juruq","keywords":"","datePublished":"2022-09-16"}
          }
        }
      ],
      "memo": "",
      "timeout_height": "0",
      "extension_options": [
      ],
      "non_critical_extension_options": [
      ]
    },
    "auth_info": {
      "signer_infos": [
        {
          "public_key": {
            "@type": "/cosmos.crypto.secp256k1.PubKey",
            "key": "AkXlNHUXHqvjkzOWu+hdzfaDJnAgkIne8kKAAxV9b+Qj"
          },
          "mode_info": {
            "single": {
              "mode": "SIGN_MODE_DIRECT"
            }
          },
          "sequence": "10593"
        }
      ],
      "fee": {
        "amount": [
          {
            "denom": "nanolike",
            "amount": "1941470"
          }
        ],
        "gas_limit": "194147",
        "payer": "",
        "granter": ""
      }
    },
    "signatures": [
      "pPHbZpJkrp0Janoxuecogrwb3gXOuEGvrl6pARnNFu0D61fZ5UZX78+jdm3T+Yk8yStdngZyCclwJo/0db8+lw=="
    ]
  },
  "tx_response": {
    "height": "5634665",
    "txhash": "BE1F9DBF9F70B1F5E207F98806DA9B10419776A338F1707D28A07A46E10E978E",
    "codespace": "",
    "code": 0,
    "data": "0AAC010A232F6C696B65636861696E2E6973636E2E4D73674372656174654973636E5265636F72641284010A436973636E3A2F2F6C696B65636F696E2D636861696E2F4D6B685231576B6737476E4167683946497667366E5952374E46534255343378656C7944534E6B6A724A452F31123D6261677571656572613275767067656C666D73716D6E6479323277357473346E763673736B6F3734326F626B6762667262647373726674766469363571",
    "raw_log": "[{\"events\":[{\"type\":\"coin_received\",\"attributes\":[{\"key\":\"receiver\",\"value\":\"like17xpfvakm2amg962yls6f84z3kell8c5lr9lzgx\"},{\"key\":\"amount\",\"value\":\"3534000nanolike\"}]},{\"type\":\"coin_spent\",\"attributes\":[{\"key\":\"spender\",\"value\":\"like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn\"},{\"key\":\"amount\",\"value\":\"3534000nanolike\"}]},{\"type\":\"iscn_record\",\"attributes\":[{\"key\":\"iscn_id\",\"value\":\"iscn://likecoin-chain/MkhR1Wkg7GnAgh9FIvg6nYR7NFSBU43xelyDSNkjrJE/1\"},{\"key\":\"iscn_id_prefix\",\"value\":\"iscn://likecoin-chain/MkhR1Wkg7GnAgh9FIvg6nYR7NFSBU43xelyDSNkjrJE\"},{\"key\":\"owner\",\"value\":\"like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn\"},{\"key\":\"ipld\",\"value\":\"baguqeera2uvpgelfmsqmndy22w5ts4nv6ssko742obkgbfrbdssrftvdi65q\"}]},{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"/likechain.iscn.MsgCreateIscnRecord\"},{\"key\":\"sender\",\"value\":\"like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn\"},{\"key\":\"module\",\"value\":\"iscn\"},{\"key\":\"sender\",\"value\":\"like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn\"}]},{\"type\":\"transfer\",\"attributes\":[{\"key\":\"recipient\",\"value\":\"like17xpfvakm2amg962yls6f84z3kell8c5lr9lzgx\"},{\"key\":\"sender\",\"value\":\"like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn\"},{\"key\":\"amount\",\"value\":\"3534000nanolike\"}]}]}]",
    "logs": [
      {
        "msg_index": 0,
        "log": "",
        "events": [
          {
            "type": "coin_received",
            "attributes": [
              {
                "key": "receiver",
                "value": "like17xpfvakm2amg962yls6f84z3kell8c5lr9lzgx"
              },
              {
                "key": "amount",
                "value": "3534000nanolike"
              }
            ]
          },
          {
            "type": "coin_spent",
            "attributes": [
              {
                "key": "spender",
                "value": "like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn"
              },
              {
                "key": "amount",
                "value": "3534000nanolike"
              }
            ]
          },
          {
            "type": "iscn_record",
            "attributes": [
              {
                "key": "iscn_id",
                "value": "iscn://likecoin-chain/MkhR1Wkg7GnAgh9FIvg6nYR7NFSBU43xelyDSNkjrJE/1"
              },
              {
                "key": "iscn_id_prefix",
                "value": "iscn://likecoin-chain/MkhR1Wkg7GnAgh9FIvg6nYR7NFSBU43xelyDSNkjrJE"
              },
              {
                "key": "owner",
                "value": "like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn"
              },
              {
                "key": "ipld",
                "value": "baguqeera2uvpgelfmsqmndy22w5ts4nv6ssko742obkgbfrbdssrftvdi65q"
              }
            ]
          },
          {
            "type": "message",
            "attributes": [
              {
                "key": "action",
                "value": "/likechain.iscn.MsgCreateIscnRecord"
              },
              {
                "key": "sender",
                "value": "like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn"
              },
              {
                "key": "module",
                "value": "iscn"
              },
              {
                "key": "sender",
                "value": "like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn"
              }
            ]
          },
          {
            "type": "transfer",
            "attributes": [
              {
                "key": "recipient",
                "value": "like17xpfvakm2amg962yls6f84z3kell8c5lr9lzgx"
              },
              {
                "key": "sender",
                "value": "like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn"
              },
              {
                "key": "amount",
                "value": "3534000nanolike"
              }
            ]
          }
        ]
      }
    ],
    "info": "",
    "gas_wanted": "194147",
    "gas_used": "176561",
    "tx": {
      "@type": "/cosmos.tx.v1beta1.Tx",
      "body": {
        "messages": [
          {
            "@type": "/likechain.iscn.MsgCreateIscnRecord",
            "from": "like18tu779r87ezhavtuw6x03m3th2l9jas0uru8yn",
            "record": {
              "recordNotes": "like13f30lx5q8dscgxpdfpqkakx8ejwgstxqgda0s6",
              "contentFingerprints": [
                "hash://sha256/bafyreia4rcfbi3heobobu7sysnyfvtwplt5wgooxe2spqexkhijs4juruq",
                "ipfs://QmRrVSxdrCRSQ5jb5A7BCY8aReFMqhtUzuaicnE5C7vZbP"
              ],
              "stakeholders": [
                {"entity":{"@id":"did:cosmos:cosmos13f30lx5q8dscgxpdfpqkakx8ejwgstxqm3pdnp","name":"困天索 (@jerryokla198)"},"rewardProportion":100,"contributionType":"http://schema.org/author"},
                {"rewardProportion":0,"contributionType":"http://schema.org/publisher","entity":{"name":"Matters.News"}}
              ],
              "contentMetadata": {"@context":"http://schema.org/","@type":"Article","name":"分享一個可以做且會賺錢的生意","description":"大家好！跟大家分享一個近3年來很罕見 我認為可以做且會賺錢的生意！這是結合酒店、經紀傳播、博弈及超跑租賃業，所整合的一個「創新電子化高端酒店」的生意  目前有許多同業都非常認同這個案子  我認為 它不只是 台灣第一  甚至是     全球第一  如果有錢希望大家可以了解看看  \ud83d...","version":1,"url":"https://matters.news/@jerryokla198/328574-分享一個可以做且會賺錢的生意-bafyreia4rcfbi3heobobu7sysnyfvtwplt5wgooxe2spqexkhijs4juruq","keywords":"","datePublished":"2022-09-16"}
            }
          }
        ],
        "memo": "",
        "timeout_height": "0",
        "extension_options": [
        ],
        "non_critical_extension_options": [
        ]
      },
      "auth_info": {
        "signer_infos": [
          {
            "public_key": {
              "@type": "/cosmos.crypto.secp256k1.PubKey",
              "key": "AkXlNHUXHqvjkzOWu+hdzfaDJnAgkIne8kKAAxV9b+Qj"
            },
            "mode_info": {
              "single": {
                "mode": "SIGN_MODE_DIRECT"
              }
            },
            "sequence": "10593"
          }
        ],
        "fee": {
          "amount": [
            {
              "denom": "nanolike",
              "amount": "1941470"
            }
          ],
          "gas_limit": "194147",
          "payer": "",
          "granter": ""
        }
      },
      "signatures": [
        "pPHbZpJkrp0Janoxuecogrwb3gXOuEGvrl6pARnNFu0D61fZ5UZX78+jdm3T+Yk8yStdngZyCclwJo/0db8+lw=="
      ]
    },
    "timestamp": "2022-09-16T16:30:32Z",
    "events": [
      {
        "type": "coin_spent",
        "attributes": [
          {
            "key": "c3BlbmRlcg==",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bg==",
            "index": true
          },
          {
            "key": "YW1vdW50",
            "value": "MTk0MTQ3MG5hbm9saWtl",
            "index": true
          }
        ]
      },
      {
        "type": "coin_received",
        "attributes": [
          {
            "key": "cmVjZWl2ZXI=",
            "value": "bGlrZTE3eHBmdmFrbTJhbWc5NjJ5bHM2Zjg0ejNrZWxsOGM1bHI5bHpneA==",
            "index": true
          },
          {
            "key": "YW1vdW50",
            "value": "MTk0MTQ3MG5hbm9saWtl",
            "index": true
          }
        ]
      },
      {
        "type": "transfer",
        "attributes": [
          {
            "key": "cmVjaXBpZW50",
            "value": "bGlrZTE3eHBmdmFrbTJhbWc5NjJ5bHM2Zjg0ejNrZWxsOGM1bHI5bHpneA==",
            "index": true
          },
          {
            "key": "c2VuZGVy",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bg==",
            "index": true
          },
          {
            "key": "YW1vdW50",
            "value": "MTk0MTQ3MG5hbm9saWtl",
            "index": true
          }
        ]
      },
      {
        "type": "message",
        "attributes": [
          {
            "key": "c2VuZGVy",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bg==",
            "index": true
          }
        ]
      },
      {
        "type": "tx",
        "attributes": [
          {
            "key": "ZmVl",
            "value": "MTk0MTQ3MG5hbm9saWtl",
            "index": true
          }
        ]
      },
      {
        "type": "tx",
        "attributes": [
          {
            "key": "YWNjX3NlcQ==",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bi8xMDU5Mw==",
            "index": true
          }
        ]
      },
      {
        "type": "tx",
        "attributes": [
          {
            "key": "c2lnbmF0dXJl",
            "value": "cFBIYlpwSmtycDBKYW5veHVlY29ncndiM2dYT3VFR3ZybDZwQVJuTkZ1MEQ2MWZaNVVaWDc4K2pkbTNUK1lrOHlTdGRuZ1p5Q2Nsd0pvLzBkYjgrbHc9PQ==",
            "index": true
          }
        ]
      },
      {
        "type": "message",
        "attributes": [
          {
            "key": "YWN0aW9u",
            "value": "L2xpa2VjaGFpbi5pc2NuLk1zZ0NyZWF0ZUlzY25SZWNvcmQ=",
            "index": true
          }
        ]
      },
      {
        "type": "coin_spent",
        "attributes": [
          {
            "key": "c3BlbmRlcg==",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bg==",
            "index": true
          },
          {
            "key": "YW1vdW50",
            "value": "MzUzNDAwMG5hbm9saWtl",
            "index": true
          }
        ]
      },
      {
        "type": "coin_received",
        "attributes": [
          {
            "key": "cmVjZWl2ZXI=",
            "value": "bGlrZTE3eHBmdmFrbTJhbWc5NjJ5bHM2Zjg0ejNrZWxsOGM1bHI5bHpneA==",
            "index": true
          },
          {
            "key": "YW1vdW50",
            "value": "MzUzNDAwMG5hbm9saWtl",
            "index": true
          }
        ]
      },
      {
        "type": "transfer",
        "attributes": [
          {
            "key": "cmVjaXBpZW50",
            "value": "bGlrZTE3eHBmdmFrbTJhbWc5NjJ5bHM2Zjg0ejNrZWxsOGM1bHI5bHpneA==",
            "index": true
          },
          {
            "key": "c2VuZGVy",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bg==",
            "index": true
          },
          {
            "key": "YW1vdW50",
            "value": "MzUzNDAwMG5hbm9saWtl",
            "index": true
          }
        ]
      },
      {
        "type": "message",
        "attributes": [
          {
            "key": "c2VuZGVy",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bg==",
            "index": true
          }
        ]
      },
      {
        "type": "iscn_record",
        "attributes": [
          {
            "key": "aXNjbl9pZA==",
            "value": "aXNjbjovL2xpa2Vjb2luLWNoYWluL01raFIxV2tnN0duQWdoOUZJdmc2bllSN05GU0JVNDN4ZWx5RFNOa2pySkUvMQ==",
            "index": true
          },
          {
            "key": "aXNjbl9pZF9wcmVmaXg=",
            "value": "aXNjbjovL2xpa2Vjb2luLWNoYWluL01raFIxV2tnN0duQWdoOUZJdmc2bllSN05GU0JVNDN4ZWx5RFNOa2pySkU=",
            "index": true
          },
          {
            "key": "b3duZXI=",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bg==",
            "index": true
          },
          {
            "key": "aXBsZA==",
            "value": "YmFndXFlZXJhMnV2cGdlbGZtc3FtbmR5MjJ3NXRzNG52NnNza283NDJvYmtnYmZyYmRzc3JmdHZkaTY1cQ==",
            "index": true
          }
        ]
      },
      {
        "type": "message",
        "attributes": [
          {
            "key": "bW9kdWxl",
            "value": "aXNjbg==",
            "index": true
          },
          {
            "key": "c2VuZGVy",
            "value": "bGlrZTE4dHU3NzlyODdlemhhdnR1dzZ4MDNtM3RoMmw5amFzMHVydTh5bg==",
            "index": true
          }
        ]
      }
    ]
  }
}
`
	output2 := SanitizeJSON([]byte(input2))
	expected2 := []byte(strings.ReplaceAll(input2, `\ud83d`, ``))
	if !bytes.Equal(output2, expected2) {
		t.Errorf("sanitizeJSON failed, expect %#v, got %#v", string(expected), string(output))
	}
}
