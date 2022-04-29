#!/bin/sh

# ID="John Perry Barlow"
ID="Apple Daily"
EVENT='{}'
QUERY='{"stakeholders": [{"entity": {"id": "John Perry Barlow"}}]}'
KEYWORDS='{}'
X=1000000
Y=2000000
# set enable_indexscan = off;

psql mydb << SQL
EXPLAIN ANALYZE SELECT id, tx #> '{"tx", "body", "messages", 0, "record"}' as data, events, tx #> '{"timestamp"}' as timestamp
FROM txs
WHERE events @> '$EVENT'
AND ('$QUERY' = '{}' OR tx #> '{tx, body, messages, 0, record}' @> '$QUERY'::jsonb)
AND string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') @> '$KEYWORDS'
ORDER BY ID DESC
LIMIT 10
OFFSET 100
SQL
