CREATE TYPE class_parent_type AS ENUM ('UNKNOWN', 'ISCN', 'ACCOUNT');
CREATE TABLE IF NOT EXISTS nft_class (
	id bigserial primary key,
    class_id varchar(80) primary key,     -- normally 66
    parent_type class_parent_type,
    parent_iscn_id_prefix varchar(80),
    parent_account varchar(50),
    name varchar(256),
    symbol varchar(20),
    description text,
    uri varchar(256),
    uri_hash varchar(256),
    metadata jsonb,
    config jsonb,
    price int
);

CREATE TABLE IF NOT EXISTS nft (
    id bigserial primary key,
    class_id varchar(80) references nft_class(id),
    owner varchar(50),
    nft_id varchar(80),
    uri varchar(256),
    uri_hash varchar(256),
    metadata jsonb,
	UNIQUE(class_id, nft_id)
);

CREATE TABLE IF NOT EXISTS nft_events (
    id bigserial primary key,
    height bigint,
    class_id varchar references nft_class(id),
    events varchar[],
    tx jsonb
);
