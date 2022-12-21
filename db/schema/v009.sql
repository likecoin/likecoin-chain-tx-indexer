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
