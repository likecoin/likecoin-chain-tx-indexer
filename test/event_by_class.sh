#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8998"
[ -z $CLASS ] && CLASS="likenft1ewcmlx4kq4uwctmz08zs6kh5s9te4g6nghrz8rwdharpd6wsxe4s0zgjlj"
req="$HOST/likechain/likenft/v1/event?class_id=$CLASS&verbose=$VERBOSE"
echo $req
curl $req | jq
