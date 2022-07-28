#/bin/bash

BEGIN=1995
END=0

psql <<SQL
select id, count(*) OVER ()
from iscn
WHERE ('$BEGIN' = 0 OR id > '$BEGIN')
AND ('$END' = 0 OR id < '$END')
order by id desc
limit 10;
SQL
