SELECT i.owner as author, array_agg(DISTINCT n.owner) as collector
FROM iscn as i
JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
JOIN nft AS n ON c.class_id = n.class_id
GROUP BY i.owner
