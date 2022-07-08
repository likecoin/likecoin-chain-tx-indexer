#!/bin/sh

[ -z $1 ] && OWNER="like1th426dy3wnu3aeqkz7efmkfalh9w7gwkvtl567" || OWNER=$1
psql nftdev <<SQL
SELECT n.nft_id, n.class_id, n.uri, n.uri_hash, n.metadata
FROM nft as n
JOIN nft_class as c
ON n.class_id = c.class_id
WHERE owner = '$OWNER'
SQL

