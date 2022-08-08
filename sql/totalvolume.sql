CREATE INDEX IF NOT EXISTS idx_first_message ON txs USING GIN ((tx #> '{"tx", "body", "messages", 0}') jsonb_path_ops);
select sum((tx #>> '{"tx", "body", "messages", 0, "msgs", 0, "amount", 0, "amount"}')::bigint)
    from txs
WHERE tx #> '{"tx", "body", "messages", 0}' @> '{"@type": "/cosmos.authz.v1beta1.MsgExec","grantee": "like17m4vwrnhjmd20uu7tst7nv0kap6ee7js69jfrs"}' 
