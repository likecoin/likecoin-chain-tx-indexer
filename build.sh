#!/bin/bash

set -e

pushd "$(dirname "$0")" > /dev/null
docker build -t likechain/tx-indexer .
popd > /dev/null
