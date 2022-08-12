AFTER=$(date -d "2022-08-01" +%s)
BEFORE=$(date -d "2022-08-05" +%s)
req="likechain/likenft/v1/ranking?after=$AFTER&before=$BEFORE"
