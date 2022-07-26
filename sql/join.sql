SELECT *
FROM iscn as i
JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
JOIN nft_event as e ON c.class_id = e.class_id AND e.action = 'new_class'
ORDER BY i.id
LIMIT 20
