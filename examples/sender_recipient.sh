#!/bin/bash
ENDPOINT=http://localhost:8997/cosmos/tx/v1beta1/txs
curl "$ENDPOINT?events=transfer.sender='cosmos1w4hq98jtjg729ft4um63y7z4l9wdtgrlv9n5y0'&events=transfer.recipient='cosmos18ty9kzlnjgh3rcqlvkwarn9k8upcnsx89d4fmy'" | jq