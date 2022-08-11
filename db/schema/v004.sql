CREATE INDEX IF NOT EXISTS idx_first_message ON txs USING GIN ((tx #> '{"tx", "body", "messages", 0}') jsonb_path_ops);

UPDATE meta SET height = 4 WHERE id = 'schema_version';
