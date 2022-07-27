#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8999"
[ -z $LIMIT ] && LIMIT=10

# BEFORE=$(date -d "2022-07-01" +%s)

req="$HOST/likechain/likenft/v1/ranking?after=$AFTER&before=$BEFORE&limit=$LIMIT"
echo $req
curl $req | jq
