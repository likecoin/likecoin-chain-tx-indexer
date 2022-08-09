-- Version 0 is a chaotic state.
-- It could be either a completely new database, or table structures exist but
-- with no explicit version (i.e. no `schema_version` row in `meta` table)
-- So we need thing like `CREATE TABLE IF NOT EXISTS` to ensure duplicated
-- migration works for version 0 -> 1.
-- This is not the case for latter versions.

CREATE TABLE IF NOT EXISTS txs (
  id BIGSERIAL PRIMARY KEY,
  height BIGINT,
  tx_index INT,
  tx JSONB,
  events VARCHAR ARRAY,
  UNIQUE (height, tx_index)
);

CREATE TABLE IF NOT EXISTS iscn (
  id BIGSERIAL PRIMARY KEY,
  iscn_id TEXT,
  iscn_id_prefix TEXT,
  version INT,
  owner TEXT,
  name TEXT,
  description TEXT,
  url TEXT,
  keywords TEXT[],
  fingerprints TEXT[],
  ipld TEXT,
  timestamp TIMESTAMP,
  stakeholders JSONB,
  data JSONB,
  UNIQUE(iscn_id)
);

CREATE TABLE IF NOT EXISTS meta (
  id TEXT PRIMARY KEY,
  height BIGINT
);

INSERT INTO meta VALUES ('extractor_v1', 0)
ON CONFLICT DO NOTHING;

DO $$ BEGIN
  CREATE TYPE class_parent_type AS ENUM ('UNKNOWN', 'ISCN', 'ACCOUNT');
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS nft_class (
  id BIGSERIAL PRIMARY KEY,
  class_id TEXT UNIQUE,
  parent_type class_parent_type,
  parent_iscn_id_prefix TEXT,
  parent_account TEXT,
  name TEXT,
  symbol TEXT,
  description TEXT,
  uri TEXT,
  uri_hash TEXT,
  metadata JSONB,
  config JSONB,
  price INT,
  created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS nft (
  id BIGSERIAL PRIMARY KEY,
  class_id TEXT,
  owner TEXT,
  nft_id TEXT,
  uri TEXT,
  uri_hash TEXT,
  metadata JSONB,
  UNIQUE(class_id, nft_id)
);

CREATE TABLE IF NOT EXISTS nft_event (
  id BIGSERIAL PRIMARY KEY,
  action TEXT,
  class_id TEXT,
  nft_id TEXT,
  sender TEXT,
  receiver TEXT,
  events TEXT[],
  tx_hash TEXT,
  timestamp TIMESTAMP,
  UNIQUE(action, class_id, nft_id, tx_hash)
);

CREATE INDEX IF NOT EXISTS idx_txs_txhash ON txs USING HASH ((tx->>'txhash'));

CREATE INDEX IF NOT EXISTS idx_txs_height_tx_index ON txs (height, tx_index);

CREATE INDEX IF NOT EXISTS idx_tx_events ON txs USING GIN (events);

-- deprecated indexes, replaced by the `iscn` table
DROP INDEX IF EXISTS idx_record;
DROP INDEX IF EXISTS idx_keywords;

INSERT INTO meta VALUES ('schema_version', 1)
ON CONFLICT (id) DO UPDATE SET height = 1;
