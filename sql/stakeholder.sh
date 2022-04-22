#!/bin/sh

# ID="John Perry Barlow"
ID="Apple Daily"
EVENT='{}'
QUERY='{"stakeholders": [{"entity": {"id": "John Perry Barlow"}}]}'
KEYWORDS='{}'

psql mydb << SQL
SELECT tx #> '{"tx", "body", "messages", 0, "record"}' as data, events, tx #> '{"timestamp"}'
FROM txs
WHERE events @> '$EVENT'
AND ('$QUERY' = '{}' OR tx #> '{tx, body, messages, 0, record}' @> '$QUERY'::jsonb)
AND string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') @> '$KEYWORDS'
ORDER BY id DESC
OFFSET 0
LIMIT 10;
SQL
