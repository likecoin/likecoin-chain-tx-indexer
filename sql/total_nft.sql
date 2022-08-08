CREATE INDEX IF NOT EXISTS idx_nft_owner ON nft (owner);
CREATE INDEX IF NOT EXISTS idx_iscn_iscn_id_prefix ON iscn (iscn_id_prefix);
CREATE INDEX IF NOT EXISTS idx_nft_class_iscn_id_prefix ON nft_class (parent_iscn_id_prefix);
SELECT COUNT(n.id) FROM nft as n
JOIN nft_class AS c USING (class_id)
JOIN iscn AS i ON i.iscn_id_prefix = c.parent_iscn_id_prefix
WHERE n.owner != ALL(ARRAY[i.owner, 'like17m4vwrnhjmd20uu7tst7nv0kap6ee7js69jfrs'])
