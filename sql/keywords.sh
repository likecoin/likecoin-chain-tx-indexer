psql mydb <<SQL
EXPLAIN SELECT id, string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',')
FROM txs
WHERE string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') && '{"blockchain"}';
SELECT id, string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',')
FROM txs
WHERE string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') <> '{}'
LIMIT 100;
SQL
