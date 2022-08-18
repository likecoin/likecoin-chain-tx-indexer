# ISCN API Examples

```bash
export ENDPOINT="http://localhost:8997" 
```

* current support keys: iscn_id, owner, fingerprint, keywords
* support pagination: limit, page
  * default: limit 1, page 1

## Query by ISCN ID

```bash
# By ID
curl $ENDPOINT/iscn/records?iscn_id=iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1
# By Prefix
curl $ENDPOINT/iscn/records?iscn_id_prefix=iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY
```

## Query by owner

```bash
curl $ENDPOINT/iscn/records?owner=cosmos18q3dzavq7c6njw92344rf8ejpyqxqwzvy7ef50&limit=5
```

## Query by fingerprint

```bash
curl $ENDPOINT/iscn/records?fingerprint=ipfs://QmRzbij1C7224PNiw4cNBt1NzH7SbArkGjJGVb3y4Xpiw8
```

## Query by keyword

```bash
curl $ENDPOINT/iscn/records?keyword=decentralizehk&keyword=DAO
```

## Query by stakeholders ID

```bash
curl $ENDPOINT/iscn/records?stakeholder.id=did:cosmos:1vvxaklu364sejxe9tdwkg87aanejf8v6mwdu82
```

## Query by stakeholders name

```bash
curl $ENDPOINT/iscn/records?stakeholder.name=joshkiu
```

## Compound query

```bash
curl $ENDPOINT/iscn/records?owner=cosmos1ykkpc0dnetfsya88f5nrdd7p57kplaw8sva6pj&keyword=香港&limit=5&page=2
```

## Example response

```json
{
  "records": [
    {
      "ipld": "baguqeerayhsd7mbhbauudoftlcjqejy2k6zl4xurku3z45sq3hg7ekics5ha",
      "data": {
        "@id": "iscn://likecoin-chain/TPtbTpMco5zNmhGCX2U3UCFe4d415eyrXTabYZGm9PE/1",
        "recordTimestamp": "2021-08-19T04:31:03Z",
        "owner": "cosmos13f4glvg80zvfrrs7utft5p68pct4mcq7cgpf2p",
        "recordNotes": "",
        "contentFingerprints": [
          "ipfs://QmUnGEozva55C9Z1MLpULh6UXrUDe4Yo3g1K4ZN6ZqzyEs"
        ],
        "contentMetadata": {
          "@type": "Article",
          "name": "decentralize：無大台，有共識",
          "version": 1,
          "@context": "http://schema.org/",
          "keywords": "blockchain,DAO,decentralization,decentralizehk",
          "usageInfo": "ipfs://QmRvpQiiLA8ttSLAXEd5RArmXeG4qWEsKPmrB7KeiLSuE4"
        },
        "stakeholders": [
          {
            "entity": {
              "id": "https://like.co/undefined",
              "name": "kin ko"
            },
            "contributionType": "http://schema.org/author",
            "rewardProportion": 1
          },
          {
            "entity": {
              "id": "https://matters.news/",
              "name": "Matters",
              "description": "Matters is a decentralized, cryptocurrency driven content creation and discussion platform."
            },
            "contributionType": "http://schema.org/publisher",
            "rewardProportion": 1
          }
        ]
      }
    }
  ],
  "pagination": {
    "next_key": 2,
    "count": 1
  }
}
```
