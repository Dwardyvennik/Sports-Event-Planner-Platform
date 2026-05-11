SERVICE := api-gateway
GO_CACHE ?= /tmp/go-build

.PHONY: run test build

run:
	GOCACHE=$(GO_CACHE) go run -buildvcs=false ./cmd/$(SERVICE)

test:
	GOCACHE=$(GO_CACHE) go test -buildvcs=false ./...

build:
	GOCACHE=$(GO_CACHE) go build -buildvcs=false -o ../../bin/$(SERVICE) ./cmd/$(SERVICE)
