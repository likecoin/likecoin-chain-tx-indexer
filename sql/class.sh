#!/bin/sh

# ID="John Perry Barlow"
# TYPE='CreativeWork'
CREATOR='like1ruydfsjdqlx4m3yy0tmv89au9jzquk6xas04xp'
AFTER='2022/07/15'
BEFORE='2022/07/26'
STAKEHOLDERS='[]'
# STAKEHOLDERS='[{"name": "Carlos Cuesta"}]'

psql testnet5 << SQL
SELECT * 
FROM nft_class as c
JOIN (
    SELECT c.id
    FROM iscn as i
    JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
    JOIN nft as n ON c.class_id = n.class_id
	JOIN nft_event as e ON c.class_id = e.class_id AND e.action = 'new_class'
    WHERE ('$CREATOR' = '' OR i.owner = '$CREATOR')
        AND ('$TYPE' = '' OR i.data #>> '{"contentMetadata", "@type"}' = '$TYPE')
        AND ('$STAKEHOLDERS'::jsonb IS NULL OR i.stakeholders @> '$STAKEHOLDERS')
        AND ('$OWNER' = '' OR n.owner = '$OWNER')
		AND ('$AFTER' = '0001-01-01T00:00:00Z' OR e.timestamp > '$AFTER')
		AND ('$BEFORE' = '0001-01-01T00:00:00Z' OR e.timestamp < '$BEFORE')
    GROUP BY c.id
) as t
ON c.id = t.id
SQL
