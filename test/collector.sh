#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8999"
[ -z $ISCN ] && AUTHOR="like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf"
[ -z $LIMIT ] && LIMIT=5
[ -z $OFFSET ] && OFFSET=0
req="$HOST/likechain/likenft/v1/collector?creator=$AUTHOR&limit=$LIMIT&offset=$OFFSET&reverse=true"
echo $req
curl $req | jq
