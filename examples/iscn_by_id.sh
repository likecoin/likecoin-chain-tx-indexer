#/bin/bash
ISCN=$1
if [[ -z $ISCN ]]; then
    ISCN="iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1"
fi
curl "http://localhost:8997/iscn/records/id?iscn_id=$ISCN" | jq