#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8998"
[ -z $CLASS ] && CLASS="likenft1ltlz9q5c0xu2xtrjudrgm4emfu37du755kytk8swu4s6yjm268msp6mgf8"
[ -z $NFT ] && NFT="testing-aurora-86"
req="$HOST/likenft/event?class_id=$CLASS&nft_id=$NFT&verbose=$VERBOSE"
echo $req
curl $req | jq
