#!/bin/sh

# The main file to run tests

[ -z $HOST ] && HOST="http://localhost:8999"
[ -z $LIMIT ] && LIMIT=10

FAILED=""

for t in $@; do
    echo $t
    . "./$t" &&
    req="$HOST/$req&limit=$LIMIT" &&
    echo $req &&
    curl --fail-with-body $req | jq 
done

