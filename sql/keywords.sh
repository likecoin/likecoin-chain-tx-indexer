#/bin/sh

KEYWORD=$1
[ -z $1 ] && KEYWORD='LikeCoin'
psql <<SQL
SELECT *
FROM iscn
WHERE keywords @> '{"$KEYWORD"}';
SQL
