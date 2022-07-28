CREATE TEMP TABLE IF NOT EXISTS iscn (
	id bigserial primary key,
	iscn_id text,
	iscn_id_prefix text,
	version int,
	owner text,
	name text,
	description text,
	url text,
	keywords text[],
	fingerprints text[],
	ipld text,
	timestamp timestamp,
	stakeholders jsonb,
	data jsonb,
	UNIQUE(iscn_id)
);

CREATE TEMP TABLE IF NOT EXISTS meta (
    id VARCHAR(10) PRIMARY KEY,
    height BIGINT
);

INSERT INTO iscn (iscn_id, owner, keywords, fingerprints, stakeholders, data) VALUES
(
    'iscn://likecoin-chain/NpoNU1609xrsPS4gxFsALLYoEff4b0JMQSxN-MuI3VE/1',
    'like18lyaxdl7x6fc6slp77248vvuej936779jpmxa2',
    ARRAY['test', 'LikeCoin'],
    ARRAY['https://depub.blog', 'hash://sha256/bc7c83785cc93fc79b5b4a0a24420df6e837e53f0d294929f19fde5a3abda028'],
	'[{"id": "cosmos18lyaxdl7x6fc6slp77248vvuej936779pa8y73", "name": "Justin Lin"},
		{"id": "depub.space", "name": "depub.space"}]',
    '{"contentMetadata": {"name": "hello"}}'
);

SELECT * FROM iscn;

INSERT INTO meta VALUES
    ('iscn', 257374),
    ('txs', 7773482);

UPDATE meta
    SET height = 400000
    WHERE id = 'iscn';

SELECT * FROM meta;

SELECT id, height, tx #> '{"tx", "body", "messages", 0, "record"}' as records, events, tx #> '{"timestamp"}'
FROM txs
WHERE height > (SELECT height FROM meta WHERE id = 'iscn')
	AND height < (SELECT height FROM meta WHERE id = 'iscn') + 10000
	AND events && '{"message.module=\"iscn\""}'
ORDER BY id ASC;
