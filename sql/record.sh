#/bin/bash

# FOOTPRINT="hash://sha256/d2a92fe4b7c5b9654f8aa303bed0b727931ab44c7f29b2750580abca2cb6597d"
ID="cosmos1udm9fntn8vsg7ujeznjdx8nhvx5t4rhx4fp3ra"
FOOTPRINT="ipfs://QmQTKptHHUJ8cQQfm42epks8Ty3wUPKYz8KhhvNT2z32tM"

# psql mydb <<SQL
# CREATE INDEX ON txs((tx #> '{"tx", "body", "messages", 0, "record", "contentFingerprints"}'));
# SQL

psql mydb <<SQL
select tx #> '{tx, body, messages, 0, record}'
from txs
where tx #> '{tx, body, messages, 0, record}' @> '{"stakeholders": [{"entity": {"@id": "$ID"}}]}'
SQL