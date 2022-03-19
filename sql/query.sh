#!/bin/bash

# Create index to enhance performance
# CREATE INDEX txs USING GIN(jsonb_to_tsvector('english', tx #> '{"tx", "body", "messages", 0, "record", "contentMetadata"}' , '["string"]'));

# Usage:
# ./query.sh [search term]
# Example:
# ./query.sh decentralizehk
# Multiple terms:
# ./query.sh 'decentralizehk & likecoin'


TERM=$@
psql mydb <<SQL
select tx #> '{"tx", "body", "messages", 0, "record", "contentMetadata"}'
  from txs
  where jsonb_to_tsvector('english', tx #> '{"tx", "body", "messages", 0, "record", "contentMetadata"}' , '["string"]') @@ to_tsquery('English', '$TERM')
  limit 1000
SQL
