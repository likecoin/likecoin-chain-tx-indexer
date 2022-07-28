CREATE TYPE class_parent_type AS ENUM ('UNKNOWN', 'ISCN', 'ACCOUNT');
CREATE TABLE IF NOT EXISTS nft_class (
	id bigserial primary key,
    class_id text primary key,     -- normally 66
    parent_type class_parent_type,
    parent_iscn_id_prefix text,
    parent_account text,
    name text,
    symbol text,
    description text,
    uri text,
    uri_hash text,
    metadata jsonb,
    config jsonb,
    price int,
	created_at timestamp
);

CREATE TABLE IF NOT EXISTS nft (
    id bigserial primary key,
    class_id text,
    owner text,
    nft_id text,
    uri text,
    uri_hash text,
    metadata jsonb,
	UNIQUE(class_id, nft_id)
);

CREATE TABLE IF NOT EXISTS nft_event (
    id bigserial primary key,
    action text,
    class_id text references nft_class(id),
    nft_id text,
    sender text,
    receiver text,
    events text[],
    tx_hash text,
    timestamp timestamp
);
