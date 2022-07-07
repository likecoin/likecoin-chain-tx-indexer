#!/bin/sh

[ -z $1 ] && ISCN="iscn://likecoin-chain/fIaP4-pj5cdfstg-DsE4_QEMNmzm42PS0uGQ-nPuc_Q" || ISCN=$1
echo $ISCN
psql nftdev <<SQL
SELECT c.class_id, array_agg(row_to_json(n.*))
FROM nft_class as c
LEFT JOIN nft as n ON n.class_id = c.class_id
WHERE c.parent_iscn_id_prefix = '$ISCN'
GROUP BY c.class_id
SQL

# psql nftdev <<SQL
# SELECT nft.* FROM nft_class
# LEFT JOIN nft ON nft.class_id = nft_class.class_id
# WHERE nft_class.parent_iscn_id_prefix = '$ISCN'
# SQL
