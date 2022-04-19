#!/bin/bash

psql mydb <<SQL
SELECT tx #> '{"tx", "body", "messages", 0, "record"}' AS data, events
FROM txs
WHERE events @> '{"iscn_record.iscn_id=\"iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1\""}'
SQL