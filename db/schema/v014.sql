CREATE TABLE IF NOT EXISTS nft_income (
  id BIGSERIAL PRIMARY KEY,
  class_id TEXT NOT NULL,
  nft_id TEXT NOT NULL,
  tx_hash TEXT NOT NULL,
  address TEXT NOT NULL,
  amount BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_nft_income_class_id_nft_id_tx_hash ON nft_income (class_id, nft_id, tx_hash);

CREATE INDEX IF NOT EXISTS idx_nft_income_address ON nft_income (address);