CREATE INDEX IF NOT EXISTS idx_event_receiver ON nft_event (receiver);

UPDATE meta SET height = 6 WHERE id = 'schema_version';
