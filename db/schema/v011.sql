ALTER TABLE nft
  ADD COLUMN latest_price BIGINT DEFAULT 0,
  ADD COLUMN price_updated_at timestamp DEFAULT NULL
;

ALTER TABLE nft_class
  ADD COLUMN latest_price BIGINT DEFAULT 0,
  ADD COLUMN price_updated_at timestamp DEFAULT NULL
;

-- migration is in parallel migration
