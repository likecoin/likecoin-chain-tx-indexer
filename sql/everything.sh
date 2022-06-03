#/bin/bash
TERM=$1

echo $TERM

psql mydb << SQL
set enable_indexscan = off;
select tx #> '{"tx", "body", "messages", 0, "record"}'
    from txs
    where tx #> '{tx, body, messages, 0, record}' @> '{"stakeholders": [{"entity": {"@id": "$TERM"}}]}'
            OR tx #> '{tx, body, messages, 0, record}' @> '{"stakeholders": [{"entity": {"name": "$TERM"}}]}'
            OR tx #> '{tx, body, messages, 0, record}' @> '{"contentFingerprints": ["$TERM"]}'
            OR string_to_array(tx #>> '{tx, body, messages, 0, record, contentMetadata, keywords}', ',') @> '{"$TERM"}'
            OR events @> '{"iscn_record.owner=\"$TERM\""}'
            OR events @> '{"iscn_record.iscn_id=\"$TERM\""}'
    order by id desc
    limit 10
SQL