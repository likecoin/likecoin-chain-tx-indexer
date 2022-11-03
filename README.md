# likecoin-chain-tx-indexer

This is a tool for indexing transaction from [LikeCoin chain](https://github.com/likecoin/likecoin-chain), replacing the slow `/txs?...` query endpoint provided by Tendermint and the lite client.

## Build

`go build -o indexer main.go`

For Docker image, run `./build.sh` to build and tag the Docker image.

## Usage

You need a Postgres server as the storage database of indexed transactions.

In the following commands, you can specify the Postgres connection by providing `postgres-db`, `postgres-host`, `postgres-port`, `postgres-user`, `postgres-pwd` parameters.

You may refer to `docker-compose.yml` provided for Docker setup.

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

### testing

You may run a testing Postgres database:

```
docker compose run --rm -p 127.0.0.1:5433:5432 test-db
```

Then run `go test ./...`.

Alternatively, provide an empty Postrgres database with environment variables:

```
DB_NAME=my_pg_db DB_HOST=somewhere DB_PORT=15432 DB_USER=my_pg_user DB_PASS=my_password go test ./...
```

If you really want to test on production server, you may add `TEST_ON_PRODUCTION=1` environment variable to enforce it.

### API Example

Please refer to [examples](./examples)
