#!/bin/sh

[ -z $ISCN ] && ISCN="iscn://likecoin-chain/IKI9PueuJiOsYvhN6z9jPJIm3UGMh17BQ3tEwEzslQo"
EXPAND=false
req="likechain/likenft/v1/class?iscn_id_prefix=$ISCN&expand=$EXPAND" 
