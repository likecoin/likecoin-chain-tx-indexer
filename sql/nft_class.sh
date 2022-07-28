#!/bin/sh
psql nftdev <<SQL
select tx #> '{"tx", "body", "messages"}' as messages, tx -> 'logs' as logs from txs
where events @> '{"message.action=\"new_class\""}'
limit 5
SQL
