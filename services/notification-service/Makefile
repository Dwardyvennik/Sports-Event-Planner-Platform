SERVICE := notification-service
GO_CACHE ?= /tmp/go-build

.PHONY: run test build migrate-up migrate-down

run:
	GOCACHE=$(GO_CACHE) go run -buildvcs=false ./cmd/$(SERVICE)

test:
	GOCACHE=$(GO_CACHE) go test -buildvcs=false ./...

build:
	GOCACHE=$(GO_CACHE) go build -buildvcs=false -o ../../bin/$(SERVICE) ./cmd/$(SERVICE)

migrate-up:
	cd ../.. && ./scripts/migrate.sh up

migrate-down:
	cd ../.. && ./scripts/migrate.sh down 1
