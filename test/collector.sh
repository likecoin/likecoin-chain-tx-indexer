#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8998"
[ -z $ISCN ] && AUTHOR="like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf"
[ -z $LIMIT ] && LIMIT=10
req="$HOST/likenft/collector?creator=$AUTHOR"
echo $req
curl $req | jq
