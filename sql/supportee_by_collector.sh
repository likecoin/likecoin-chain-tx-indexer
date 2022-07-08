#/bin/sh
[ -z $1 ] && OWNER="like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp" || OWNER=$1

echo "supportee of $OWNER"
psql nftdev <<SQL
SELECT i.owner as author, COUNT(*), SUM(c.price), array_agg(DISTINCT c.class_id) as collections
FROM iscn as i
JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
JOIN nft AS n ON c.class_id = n.class_id
WHERE n.owner = '$OWNER'
GROUP BY i.owner
SQL
