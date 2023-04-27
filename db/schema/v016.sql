ALTER TABLE nft_event
  ADD COLUMN iscn_owner_at_the_time TEXT DEFAULT '' NOT NULL
;

CREATE INDEX idx_nft_event_iscn_owner_at_the_time ON nft_event (iscn_owner_at_the_time);
