#!/bin/sh

# ID="John Perry Barlow"
# TYPE='CreativeWork'
# CREATOR='like1ruydfsjdqlx4m3yy0tmv89au9jzquk6xas04xp'
AFTER='2022/07/15'
BEFORE='2022/07/26'
STAKEHOLDERS='[]'
IGNORELIST="{}" # "{\"like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp\"}"
ALLOWOWNER=false
OWNER=like1yney2cqn5qdrlc50yr5l53898ufdhxafqz9gxp # like13v8qtt0jz6y2304559v7l29sy7prz50jqwdewn
# STAKEHOLDERS='[{"name": "Carlos Cuesta"}]'

psql testnet5 << SQL
SELECT c.class_id, count, owners
FROM nft_class as c
JOIN (
    SELECT c.id, count(n.id), array_agg(DISTINCT n.owner) as owners
    FROM iscn as i
    JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
    LEFT JOIN nft as n ON c.class_id = n.class_id
        AND ('$ALLOWOWNER' = true OR n.owner != i.owner)
        AND n.owner != ALL('$IGNORELIST'::text[]) 
    JOIN nft_event as e ON c.class_id = e.class_id AND e.action = 'new_class'
    WHERE ('$CREATOR' = '' OR i.owner = '$CREATOR')
        AND ('$TYPE' = '' OR i.data #>> '{"contentMetadata", "@type"}' = '$TYPE')
        AND ('$STAKEHOLDERS'::jsonb IS NULL OR i.stakeholders @> '$STAKEHOLDERS')
        AND ('$AFTER' = '0001-01-01T00:00:00Z' OR e.timestamp > '$AFTER')
        AND ('$BEFORE' = '0001-01-01T00:00:00Z' OR e.timestamp < '$BEFORE')
    GROUP BY c.id
) as t
ON c.id = t.id
WHERE ('$OWNER' = '' OR '$OWNER' = ANY(owners))
ORDER BY count DESC
SQL
