#/bin/bash

psql mydb <<SQL
select id, tx #> '{tx, body, messages, 0, record}' as record
from txs
where events @> '{"message.module=\"iscn\""}' 
order by id desc
limit 10;
SQL