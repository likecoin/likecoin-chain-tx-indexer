#/bin/sh
# [ -z $1 ] && OWNER="like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf" || OWNER=$1
OWNER="like1qv66yzpgg9f8w46zj7gkuk9wd2nrpqmca3huxf" 
[ -z $2 ] && OFFSET="0" || OFFSET=$2
[ -z $3 ] && LIMIT="10" || LIMIT=$3

echo "supporters of $OWNER"
psql testnet5 <<SQL
SELECT owner, sum(count), 
	array_agg(json_build_object(
		'iscn_id_prefix', iscn_id_prefix, 
		'class_id', class_id, 
		'count', count))
FROM (
	SELECT n.owner, i.iscn_id_prefix, c.class_id, COUNT(*) as count
	FROM iscn as i
	JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
	JOIN nft AS n ON c.class_id = n.class_id
	WHERE i.owner = '$OWNER'
	GROUP BY n.owner, i.iscn_id_prefix, c.class_id
) as r
GROUP BY owner
OFFSET '$OFFSET'
LIMIT '$LIMIT'
SQL
