#/bin/bash

BEGIN=1995
END=2000

psql <<SQL
select id
from iscn
WHERE ('$BEGIN' = 0 OR id > '$BEGIN')
AND ('$END' = 0 OR id < '$END')
order by id desc
limit 10;
SQL
