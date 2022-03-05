#!/bin/bash

# Create index to enhance performance
# create index on txs using gin(jsonb_to_tsvector('english', tx #> '{"tx", "body", "messages", 0, "record", "contentMetadata"}' , '["string"]'));
TERM=$1
psql mydb <<SQL
select tx #> '{"tx", "body", "messages", 0, "record", "contentMetadata"}'
  from txs
  where jsonb_to_tsvector('english', tx #> '{"tx", "body", "messages", 0, "record", "contentMetadata"}' , '["string"]') @@ plainto_tsquery('English', '$TERM')
SQL
