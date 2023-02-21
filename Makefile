#!/usr/bin/make -f

COMMIT := $(shell git rev-parse HEAD)

build:
	go build -o indexer main.go

build-image:
	docker build --build-arg INDEXER_COMMIT_HASH=$(COMMIT) -t likechain/tx-indexer .

build-and-push:
	docker buildx --build-arg INDEXER_COMMIT_HASH=$(COMMIT) build -t likechain/tx-indexer:latest --platform linux/amd64 .
	docker tag likechain/tx-indexer:latest us.gcr.io/likecoin-foundation/likechain-tx-indexer:latest
	docker -- push us.gcr.io/likecoin-foundation/likechain-tx-indexer:latest

build-and-push-develop:
	docker buildx --build-arg INDEXER_COMMIT_HASH=$(COMMIT) build -t likechain/tx-indexer:latest --platform linux/amd64 .
	docker tag likechain/tx-indexer:latest us.gcr.io/likecoin-develop/likechain-tx-indexer:latest
	docker -- push us.gcr.io/likecoin-develop/likechain-tx-indexer:latest

install-tools-golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1

install-tools-goimports:
	go install golang.org/x/tools/cmd/goimports@latest

install-tools: install-tools-golangci-lint install-tools-goimports

lint:
	golangci-lint run

format:
	find . -name '*.go' -type f -not -path "*.git*" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "*.git*" | xargs goimports -w

test:
	go test --count=1 ./...

.PHONY: build build-image build-and-push install-tools-golangci-lint install-tools-goimports install-tools lint format test
