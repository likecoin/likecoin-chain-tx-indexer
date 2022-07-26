#!/bin/sh

[ -z $HOST ] && HOST="http://localhost:8999"
[ -z $LIMIT ] && LIMIT=10

req="$HOST/likenft/ranking?after=2022-07-15T00:00:00Z&before=2022-07-26T00:00:00Z&limit=$LIMIT"
echo $req
curl $req | jq

req="$HOST/likenft/ranking?creator=like1utqnsl38fuz5m0yl0u054ducla9wryltjk83h3"
echo $req
curl $req | jq

req="$HOST/likenft/ranking?type=CreativeWork"
echo $req
curl $req | jq

req="$HOST/likenft/ranking?owner=like1hdgwjdzzejavvc5zqts0aplwh7vkpj64w2cuf0"
echo $req
curl $req | jq

req="$HOST/likenft/ranking?stakeholder_id=did:like:1utqnsl38fuz5m0yl0u054ducla9wryltjk83h3"
echo $req
curl $req | jq

req="$HOST/likenft/ranking?stakeholder_name=Author"
echo $req
curl $req | jq

