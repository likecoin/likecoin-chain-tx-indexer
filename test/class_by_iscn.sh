#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8998"
[ -z $ISCN ] && ISCN="iscn://likecoin-chain/LfNpeWjRHC8NncTZNZ1pJlWDvApnESJCU9b6zVkzEq4"
req="$HOST/likenft/class?iscn_id_prefix=$ISCN&expand=$EXPAND" 
echo $req
curl $req | jq
