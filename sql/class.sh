#!/bin/sh

HOLD=

psql testnet5 <<SQL
SELECT c.* 
FROM nft_class as c
JOIN iscn as i ON c.parent_iscn_id_prefix = i.iscn_id_prefix
JOIN nft as n ON c.class_id = n.class_id

LIMIT 5
SQL
