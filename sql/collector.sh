#/bin/sh
# [ -z $1 ] && OWNER="like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf" || OWNER=$1
OWNER="like13f4glvg80zvfrrs7utft5p68pct4mcq7t5atf6" 
[ -z $2 ] && OFFSET="0" || OFFSET=$2
[ -z $3 ] && LIMIT="10" || LIMIT=$3

echo "supporters of $OWNER"
psql <<SQL
SELECT owner, sum(count) as total, 
	array_agg(json_build_object(
		'iscn_id_prefix', iscn_id_prefix, 
		'class_id', class_id, 
		'count', count))
FROM (
	SELECT n.owner, i.iscn_id_prefix, c.class_id, COUNT(DISTINCT n.id) as count
	FROM iscn as i
	JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
	JOIN nft AS n ON c.class_id = n.class_id
	WHERE i.owner = '$OWNER'
	GROUP BY n.owner, i.iscn_id_prefix, c.class_id
) as r
GROUP BY owner
ORDER BY total DESC
OFFSET '$OFFSET'
LIMIT '$LIMIT'
SQL
