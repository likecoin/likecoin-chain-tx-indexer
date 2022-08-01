#!/bin/sh

AFTER=$(date -d "2022-07-20" +%s)
BEFORE=$(date -d "2022-07-25" +%s)
req="likechain/likenft/v1/ranking?after=$AFTER&before=$BEFORE"
