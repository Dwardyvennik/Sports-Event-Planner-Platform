SHELL := /bin/sh

SERVICES := auth-service event-service notification-service api-gateway
MIGRATE_IMAGE := migrate/migrate:v4.17.1
GO_CACHE ?= /tmp/go-build

.PHONY: tidy test build docker-up docker-down docker-logs migrate-up migrate-down $(SERVICES)

tidy:
	go mod tidy

test:
	GOCACHE=$(GO_CACHE) go test -buildvcs=false ./...

build:
	@for service in $(SERVICES); do \
		echo "building $$service"; \
		GOCACHE=$(GO_CACHE) go build -buildvcs=false -o bin/$$service ./services/$$service/cmd/$$service; \
	done

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down --remove-orphans

docker-logs:
	docker compose logs -f

migrate-up:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down

auth-service:
	GOCACHE=$(GO_CACHE) go run -buildvcs=false ./services/auth-service/cmd/auth-service

event-service:
	GOCACHE=$(GO_CACHE) go run -buildvcs=false ./services/event-service/cmd/event-service

notification-service:
	GOCACHE=$(GO_CACHE) go run -buildvcs=false ./services/notification-service/cmd/notification-service

api-gateway:
	GOCACHE=$(GO_CACHE) go run -buildvcs=false ./services/api-gateway/cmd/api-gateway
