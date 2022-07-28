select * from (select unnest(events) event from txs) x where event LIKE 'message.action=\"%\"' group by event;
