FROM golang:1.19-alpine AS base

RUN apk update && apk add --no-cache build-base git bash curl linux-headers ca-certificates
RUN mkdir -p ./tx-indexer
WORKDIR /tx-indexer
COPY go.mod go.sum ./
RUN go mod download

FROM base
COPY . .
RUN go build -o /bin/tx-indexer main.go
WORKDIR /bin
