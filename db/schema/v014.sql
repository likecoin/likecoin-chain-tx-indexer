CREATE TABLE IF NOT EXISTS nft_royalty (
  id BIGSERIAL PRIMARY KEY,
  class_id TEXT NOT NULL,
  nft_id TEXT NOT NULL,
  tx_hash TEXT NOT NULL,
  stakeholder_address TEXT NOT NULL,
  royalty BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_nft_royalty_class_id_nft_id_tx_hash ON nft_royalty (class_id, nft_id, tx_hash);

CREATE INDEX IF NOT EXISTS idx_nft_royalty_stakeholder ON nft_royalty (stakeholder_address);