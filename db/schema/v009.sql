INSERT INTO meta (
  SELECT
    'latest_block_height' AS id,
    height
  FROM txs
  ORDER BY height DESC
  LIMIT 1
);
INSERT INTO meta VALUES ('latest_block_height', 0)
ON CONFLICT DO NOTHING;

INSERT INTO meta (
  SELECT
    'latest_block_time_epoch_ns' AS id,
    EXTRACT(
      EPOCH FROM (tx ->> 'timestamp')::timestamptz
    )::bigint * 1000000000 AS height
  FROM txs
  ORDER BY txs.height DESC
  LIMIT 1
);
INSERT INTO meta VALUES ('latest_block_time_epoch_ns', 0)
ON CONFLICT DO NOTHING;

CREATE TABLE nft_marketplace (
  type TEXT, -- 'listing' / 'offer'
  class_id TEXT,
  nft_id TEXT,
  creator TEXT, -- seller for listings, buyer for offers
  price BIGINT,
  expiration TIMESTAMP,
  PRIMARY KEY (type, class_id, nft_id, creator)
);

CREATE INDEX idx_nft_marketplace_class_id ON nft_marketplace (
  type,
  class_id,
  expiration,
  price
);

CREATE INDEX idx_nft_marketplace_nft_id ON nft_marketplace (
  type,
  nft_id,
  expiration,
  price
);

CREATE INDEX idx_nft_marketplace_creator ON nft_marketplace (
  type,
  creator,
  expiration,
  price
);
