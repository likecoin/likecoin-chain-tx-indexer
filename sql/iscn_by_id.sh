#!/bin/bash

psql <<SQL
SELECT events
FROM txs
WHERE events @> '{"iscn_record.iscn_id=\"iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1\""}'
SQL
