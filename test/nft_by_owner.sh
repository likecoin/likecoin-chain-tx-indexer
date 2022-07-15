#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8998"
[ -z $OWNER ] && OWNER=like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf
req="$HOST/likenft/nft?owner=$OWNER"
echo $req
curl $req | jq
