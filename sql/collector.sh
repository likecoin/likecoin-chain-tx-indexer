#/bin/sh
[ -z $1 ] && OWNER="like1lsagfzrm4gz28he4wunt63sts5xzmczw5a2m42" || OWNER=$1

echo "supporters of $OWNER"
psql nftdev <<SQL
SELECT n.owner as owner, c.class_id as class_id, COUNT(*) as count
FROM iscn as i
JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
JOIN nft AS n ON c.class_id = n.class_id
WHERE i.owner = '$OWNER'
GROUP BY n.owner, c.class_id
SQL
