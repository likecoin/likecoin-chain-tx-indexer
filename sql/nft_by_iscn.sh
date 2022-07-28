#!/bin/sh

[ -z $1 ] && ISCN="iscn://likecoin-chain/fIaP4-pj5cdfstg-DsE4_QEMNmzm42PS0uGQ-nPuc_Q" || ISCN=$1
echo $ISCN
psql nftdev <<SQL
SELECT c.*, (
	SELECT array_agg(row_to_json((nft.*)))
	FROM nft
	WHERE nft.class_id = c.class_id
	GROUP BY nft.class_id
), COUNT(*) OVER()
FROM nft_class as c
WHERE c.parent_iscn_id_prefix = '$ISCN'
LIMIT 1
SQL

