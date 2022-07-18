#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8998"
[ -z $ISCN ] && ISCN="iscn://likecoin-chain/LfNpeWjRHC8NncTZNZ1pJlWDvApnESJCU9b6zVkzEq4"
[ -z $LIMIT ] && LIMIT=10
req="$HOST/likenft/class?iscn_id_prefix=$ISCN&expand=$EXPAND&limit=$LIMIT" 
echo $req
curl $req | jq
