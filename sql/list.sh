#/bin/bash

psql <<SQL
select id, tx #> '{tx, body, messages, 0, record}' as record, events
from txs
where events @> '{"message.module=\"iscn\""}'
and not events && '{"message.action=\"create_iscn_record\"",
	"message.action=\"update_iscn_record\"",
	"message.action=\"/likechain.iscn.MsgCreateIscnRecord\"",
    "message.action=\"/likechain.iscn.MsgChangeIscnRecordOwnership\"",
    "message.action=\"msg_change_iscn_record_ownership\""
}'
order by id desc
limit 10;
SQL
