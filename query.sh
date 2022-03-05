#!/bin/bash
TERM=$1
psql mydb <<SQL
select tx #> '{"tx", "body", "messages", 0, "record", "contentMetadata"}'
  from txs
  where jsonb_to_tsvector('english', tx #> '{"tx", "body", "messages", 0, "record", "contentMetadata"}' , '["string"]') @@ plainto_tsquery('English', '$TERM')
  limit 10
SQL
