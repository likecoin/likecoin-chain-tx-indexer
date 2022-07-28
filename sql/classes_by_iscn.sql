SELECT parent_iscn_id_prefix, array_agg(class_id)
FROM nft_class
GROUP BY parent_iscn_id_prefix
