#!/bin/sh

ISCN=$1
[ -z $1 ] && ISCN='iscn://likecoin-chain/xlQLEwQFeeUPgILgK6sysDfTTu8Gz_i7ZdgMi0IgjPc/1'

psql <<SQL
SELECT *
FROM iscn
WHERE iscn_id = '$ISCN'
SQL
