CREATE INDEX IF NOT EXISTS idx_event_sender ON nft_event (sender);
CREATE INDEX IF NOT EXISTS idx_event_action ON nft_event (action);
SELECT COUNT(*), sum((tx #>> '{"tx", "body", "messages", 0, "msgs", 0, "amount", 0, "amount"}')::bigint)
FROM txs
JOIN (
    SELECT DISTINCT tx_hash FROM nft_event
    WHERE sender = 'like17m4vwrnhjmd20uu7tst7nv0kap6ee7js69jfrs'
    AND action = '/cosmos.nft.v1beta1.MsgSend'
) t
ON tx_hash = tx ->> 'txhash'::text
