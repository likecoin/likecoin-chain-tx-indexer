CREATE INDEX IF NOT EXISTS idx_iscn_id_prefix ON iscn (iscn_id_prefix);
CREATE INDEX IF NOT EXISTS idx_nft_class_iscn_prefix ON nft_class (parent_iscn_id_prefix);
CREATE INDEX IF NOT EXISTS idx_nft_owner ON nft (owner);
CREATE INDEX IF NOT EXISTS idx_nft_class_class_id ON nft_class (class_id);
CREATE INDEX IF NOT EXISTS idx_nft_class_id ON nft (class_id);

UPDATE meta SET height = 3 WHERE id = 'schema_version';
