ALTER TABLE nft_event
  ADD COLUMN memo TEXT DEFAULT '' NOT NULL
;

-- migration is in parallel migration
