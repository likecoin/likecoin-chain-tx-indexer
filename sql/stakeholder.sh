#!/bin/sh

# ID="John Perry Barlow"
ID="iscn://likecoin-chain/laa5PLHfQO2eIfiPB2-ZnFLQrmSXOgL-NvoxyBTXHvY/1"
KEYWORDS='{}'
FINGERPRINTS='{}'
STAKEHOLDERS='[]'
# set enable_indexscan = off;
'''

'''

psql << SQL
SELECT iscn_id, owner, timestamp, ipld, data
FROM iscn
WHERE	('$ID' = '' OR iscn_id = '$ID')
    AND ('$OWNER' = '' OR owner = '$OWNER')
    AND ('$KEYWORDS'::varchar[] IS NULL OR keywords @> '$KEYWORDS')
    AND ('$FINGERPRINTS'::varchar[] IS NULL OR fingerprints @> '$FINGERPRINTS')
    AND ('$STAKEHOLDERS'::jsonb IS NULL OR stakeholders @> '$STAKEHOLDERS')
ORDER BY id DESC
LIMIT 5
SQL
