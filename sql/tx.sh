#!/bin/bash

psql mydb <<SQL
select tx from txs 
where events @> '{"message.module=\"iscn\""}'
order by id desc
limit 5
SQL