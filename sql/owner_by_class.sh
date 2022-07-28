#/bin/sh
[ -z $1 ] && CLASS="likenft1ltlz9q5c0xu2xtrjudrgm4emfu37du755kytk8swu4s6yjm268msp6mgf8" || CLASS=$1

psql nftdev <<SQL
SELECT class_id, owner, array_agg(nft_id)
FROM nft
WHERE class_id = '$CLASS'
GROUP BY owner, class_id
SQL
