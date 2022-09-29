#!/bin/bash

set -e

pushd "$(dirname "$0")" > /dev/null
docker buildx build -t likechain/tx-indexer:latest --platform linux/amd64 .
docker tag likechain/tx-indexer:latest us.gcr.io/likecoin-foundation/likechain-tx-indexer:latest
docker -- push us.gcr.io/likecoin-foundation/likechain-tx-indexer:latest
popd > /dev/null
