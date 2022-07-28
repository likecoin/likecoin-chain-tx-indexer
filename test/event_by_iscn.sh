#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8998"
[ -z $ISCN ] && ISCN="iscn://likecoin-chain/LfNpeWjRHC8NncTZNZ1pJlWDvApnESJCU9b6zVkzEq4&verbose"
req="$HOST/likechain/likenft/v1/event?iscn_id_prefix=$ISCN&verbose=$VERBOSE"
echo $req
curl $req | jq
