FROM golang:1.16-alpine

RUN apk update && apk add --no-cache build-base git bash curl linux-headers ca-certificates
RUN mkdir -p ./tx-indexer
WORKDIR /tx-indexer
COPY . .
RUN go build -o /bin/tx-indexer main.go
WORKDIR /bin
