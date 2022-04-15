# ISCN API Examples

```bash
export ENDPOINT="http://localhost:8997" 
```

* current support keys: iscn_id, owner, fingerprint, keywords
* support pagination: limit, page
  * default: limit 1, page 1

## Query by ISCN ID

```bash
curl $ENDPOINT/iscn/records?iscn_id=iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1
```

## Query by owner

```bash
curl $ENDPOINT/iscn/records?owner=cosmos18q3dzavq7c6njw92344rf8ejpyqxqwzvy7ef50&limit=5
```

## Query by fingerprint

```bash
curl $ENDPOINT/iscn/records?fingerprint=ipfs://QmRzbij1C7224PNiw4cNBt1NzH7SbArkGjJGVb3y4Xpiw8
```

## Query by keywords

```bash
curl $ENDPOINT/iscn/records?keywords=decentralizehk&keywords=DAO
```

## Compound query

```bash
curl $ENDPOINT/iscn/records?owner=cosmos1ykkpc0dnetfsya88f5nrdd7p57kplaw8sva6pj&keywords=香港&limit=5&page=2
```

## Example response

```json
{
  "records": [
    {
      "ipld": "baguqeeragy7ojsthcjfsqqktqtjpmamrhzrlgtkvglucje3fptqu7thox4qa",
      "data": {
        "@context": {
          "@vocab": "http://iscn.io/",
          "contentMetadata": {
            "@context": null
          },
          "recordParentIPLD": {
            "@container": "@index"
          },
          "stakeholders": {
            "@context": {
              "@vocab": "http://schema.org/",
              "contributionType": "http://iscn.io/contributionType",
              "entity": "http://iscn.io/entity",
              "footprint": "http://iscn.io/footprint",
              "rewardProportion": "http://iscn.io/rewardProportion"
            }
          }
        },
        "@id": "iscn://likecoin-chain/cbwJejWH2jUdM4ZsP_Kh5RyN-j_SBovjieroWXY0yRQ/1",
        "@type": "Record",
        "contentFingerprints": [
          "hash://sha256/8e6984120afb8ac0f81080cf3e7a38042c472c4deb3e2588480df6e199741c89",
          "ipfs://QmVpBwQ4J1UY8tThmhRm4HqjiVPZThRyCxcN2iMi1L9rbS",
          "ar://XbVo9x7qml1-5N1wcbyLzXW-EQcCdWPq7QNmxzn7_-w"
        ],
        "contentMetadata": {
          "@context": "http://schema.org/",
          "@type": "Photo",
          "description": "狗，4歲，男性，生性貪吃",
          "exifInfo": {
            "ColorSpace": 65535,
            "ExifImageHeight": 1108,
            "ExifImageWidth": 1478,
            "ExifVersion": "2.2",
            "Format": "image/jpeg",
            "ImageUniqueID": "4e3fcc68fee70b310000000000000000",
            "Size": "1108 x 1478 JPEG (191 KB)",
            "Software": "Picasa"
          },
          "keywords": "狗,巴豆",
          "name": "巴豆",
          "url": "-",
          "usageInfo": "-",
          "version": 1
        },
        "recordNotes": "",
        "recordTimestamp": "2021-12-06T06:14:26+00:00",
        "recordVersion": 1,
        "stakeholders": [
          {
            "contributionType": "http://schema.org/author",
            "entity": {
              "@id": "did:cosmos:1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j",
              "description": "",
              "identifier": [
                {
                  "@type": "PropertyValue",
                  "propertyID": "Cosmos Wallet",
                  "value": "did:cosmos:1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmcwdt79j"
                }
              ],
              "name": "Auora Huang",
              "sameAs": [],
              "url": "https://like.co/aurorahuang22"
            },
            "rewardProportion": 1
          }
        ]
      }
    }
  ]
}
```