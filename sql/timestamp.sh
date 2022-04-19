#/bin/bash

psql mydb <<SQL
select tx #> '{"timestamp"}'
from txs
offset 2
limit 1
SQL