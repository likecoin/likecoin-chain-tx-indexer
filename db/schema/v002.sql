CREATE INDEX idx_iscn_owner ON iscn (owner);
CREATE INDEX idx_iscn_keywords ON iscn USING GIN (keywords);
CREATE INDEX idx_iscn_fingerprints ON iscn USING GIN (fingerprints);

CREATE TABLE iscn_stakeholders (
 iscn_pid BIGINT REFERENCES iscn (id),
 id TEXT,
 name TEXT
);

-- since this is a separate table, and we are usually using id field from iscn table for sorting, iscn_pid is added for the B-tree indexes
CREATE INDEX idx_iscn_stakeholders_id ON iscn_stakeholders (id, iscn_pid);
CREATE INDEX idx_iscn_stakeholders_name ON iscn_stakeholders (name, iscn_pid);

-- migration for building the iscn_stakeholders table from iscn table
INSERT INTO iscn_stakeholders (
  SELECT iscn_pid, s->>'id', s->>'name' FROM
    (SELECT id iscn_pid, jsonb_array_elements(stakeholders) s FROM iscn) t
);

INSERT INTO meta VALUES ('schema_version', 2)
ON CONFLICT (id) DO UPDATE SET height = 2;
