#!/bin/sh

[ -z $1 ] && CLASS="likenft1ltlz9q5c0xu2xtrjudrgm4emfu37du755kytk8swu4s6yjm268msp6mgf8" || CLASS=$1
[ -z $2 ] && NFT="testing-aurora-86" || NFT=$2

psql nftdev <<SQL
SELECT * 
FROM nft_event
WHERE class_id = '$CLASS' AND (nft_id = '' OR nft_id = '$NFT')
ORDER BY id
SQL
