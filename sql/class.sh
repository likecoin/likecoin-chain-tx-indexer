#!/bin/sh

# ID="John Perry Barlow"
TYPE=''
CREATOR='like1ruydfsjdqlx4m3yy0tmv89au9jzquk6xas04xp'
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
    WHERE ('$CREATOR' = '' OR i.owner = '$CREATOR')
        AND ('$TYPE' = '' OR i.data #>> '{"contentMetadata", "@type"}' = '$TYPE')
        AND ('$STAKEHOLDERS'::jsonb IS NULL OR i.stakeholders @> '$STAKEHOLDERS')
        AND ('$OWNER' = '' OR n.owner = '$OWNER')
    GROUP BY c.id
) as t
ON c.id = t.id
SQL
