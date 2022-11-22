CREATE TABLE iscn_latest_version (
  iscn_id_prefix TEXT PRIMARY KEY,
  latest_version INT
);

-- the schema_version field update is already included in db/migration.go
