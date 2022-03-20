#!/bin/bash
ENDPOINT=http://localhost:8997/cosmos/tx/v1beta1/txs
REQ="$ENDPOINT?q=web3&events=iscn_record.owner='cosmos1kcz2gaztc47zl3mcgplf9vkwl76wuzxrjvvmw5'&pagination.limit=3"
echo $REQ
curl $REQ | jq