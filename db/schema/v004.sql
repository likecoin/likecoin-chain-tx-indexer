CREATE INDEX IF NOT EXISTS idx_event_sender ON nft_event (sender);
CREATE INDEX IF NOT EXISTS idx_event_action ON nft_event (action);

UPDATE meta SET height = 4 WHERE id = 'schema_version';
