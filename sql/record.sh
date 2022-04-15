#/bin/bash

# FOOTPRINT="hash://sha256/d2a92fe4b7c5b9654f8aa303bed0b727931ab44c7f29b2750580abca2cb6597d"
ID="cosmos1udm9fntn8vsg7ujeznjdx8nhvx5t4rhx4fp3ra"
FOOTPRINT="ipfs://QmQTKptHHUJ8cQQfm42epks8Ty3wUPKYz8KhhvNT2z32tM"

# psql mydb <<SQL
# CREATE INDEX ON txs((tx #> '{"tx", "body", "messages", 0, "record", "contentFingerprints"}'));
# SQL
# create index if not exists idx_records on txs using GIN(())
# 

psql mydb <<SQL
set enable_indexscan = off;

explain select id, record
from (
    select id, jsonb_array_elements(tx #> '{tx, body, messages}') as record
    from txs
) as records
where record -> 'record' @> '{"stakeholders": [{"entity": {"@id": "$ID"}}]}'
order by id desc
limit 10;

select id, record
from (
    select id, jsonb_array_elements(tx #> '{tx, body, messages}') as record
    from txs
) as records
where record -> 'record' @> '{"stakeholders": [{"entity": {"@id": "$ID"}}]}'
order by id desc
limit 10;
SQL