# likecoin-chain-tx-indexer

This is a tool for indexing transaction from [LikeCoin chain](https://github.com/likecoin/likecoin-chain), replacing the slow `/txs?...` query endpoint provided by Tendermint and the lite client.

## Build

`go build -o indexer main.go`

For Docker image, run `./build.sh` to build and tag the Docker image.

## Usage

You need a Postgres server as the storage database of indexed transactions.

In the following commands, you can specify the Postgres connection by providing `postgres-db`, `postgres-host`, `postgres-port`, `postgres-user`, `postgres-pwd` parameters.

You may refer to the `docker-compose.yml` provided for Docker setup.

### import

```
indexer import \
    --postgres-db "postgres" \
    --postgres-host "localhost" \
    --postgres-port "5432" \
    --postgres-user "postgres" \
    --postgres-pwd "password" \
    --liked-path ".liked"
```

Import and index the transactions from existing LikeCoin chain data folder.

Note that the node needs to be shutdown before importing, since LevelDB does not allow concurrent access from different processes.

### serve

```
indexer serve \
    --postgres-db "postgres" \
    --postgres-host "localhost" \
    --postgres-port "5432" \
    --postgres-user "postgres" \
    --postgres-pwd "password" \
    --lcd-endpoint "http://localhost:1317" \
    --listen-addr ":8997"
```

Start serving the `/txs` endpoint.

The indexer will also poll and index new transactions from the lite client.

Query format is the same as the `/txs?...` endpoint of the lite client, example: `http://localhost:8997/txs?message.action=send&page=3005&limit=100`