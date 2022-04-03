#!/bin/bash
FINGERPRINT=$1
if [[ -z $FINGERPRINT ]]; then
    FINGERPRINT="ipfs://QmRzbij1C7224PNiw4cNBt1NzH7SbArkGjJGVb3y4Xpiw8"
fi
curl "http://localhost:8997/iscn/records/fingerprint?fingerprint=$FINGERPRINT" | jq