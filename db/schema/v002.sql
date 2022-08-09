CREATE INDEX idx_iscn_owner ON iscn (owner); -- 28s, 6.6s
CREATE INDEX idx_iscn_keywords ON iscn USING GIN (keywords); -- 1.7s, 2.4s
CREATE INDEX idx_iscn_fingerprints ON iscn USING GIN (fingerprints); -- 27s, 31s

CREATE TABLE iscn_stakeholders (
  -- pid means primary key ID, which is the auto-increment ID assigned by database in `iscn` table
  iscn_pid BIGINT REFERENCES iscn (id),
  -- stakeholder's ID, `sid` so not to be ambiguous with `id` column in `iscn` table
  sid TEXT,
  -- stakeholder's name, `sname` so not to be ambiguous with `name` column in `iscn` table
  sname TEXT,
  -- raw stakeholder object
  data JSONB
);

-- since this is a separate table, and we are usually using id field from iscn table for sorting, iscn_pid is added for the B-tree indexes
CREATE INDEX idx_iscn_stakeholders_sid ON iscn_stakeholders (sid, iscn_pid); -- 0.053s, 0.048s
CREATE INDEX idx_iscn_stakeholders_sname ON iscn_stakeholders (sname, iscn_pid); -- 0.049s, 0.039s

-- migration for building the iscn_stakeholders table from iscn table
INSERT INTO iscn_stakeholders (
  SELECT
    iscn_pid,
    s->>'id' AS sid,
    s->>'name' AS sname,
    obj AS data
  FROM (
    SELECT
      id AS iscn_pid,
      jsonb_array_elements(stakeholders) AS s,
      jsonb_array_elements(data#>'{stakeholders}') AS obj
    FROM iscn
  ) AS t
)
;

ALTER TABLE iscn DROP COLUMN stakeholders;

UPDATE meta SET height = 2 WHERE id = 'schema_version';
