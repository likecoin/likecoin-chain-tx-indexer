CREATE INDEX IF NOT EXISTS idx_event_nft_id ON nft_event (nft_id);
CREATE INDEX IF NOT EXISTS idx_event_nft_id_receiver ON nft_event (nft_id, receiver); -- For social graph API
