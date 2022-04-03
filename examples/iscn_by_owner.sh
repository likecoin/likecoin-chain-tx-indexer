#!/bin/bash
OWNER=$1
if [[ -z $OWNER ]]; then
    OWNER="cosmos18q3dzavq7c6njw92344rf8ejpyqxqwzvy7ef50"
fi
curl "http://localhost:8997/iscn/records/owner?owner=$OWNER" | jq