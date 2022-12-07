#!/usr/bin/make -f

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

.PHONY: install-tools-golangci-lint install-tools-goimports install-tools lint format
