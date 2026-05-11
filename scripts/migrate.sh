#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
MIGRATE_IMAGE=${MIGRATE_IMAGE:-migrate/migrate:v4.17.1}

if [ "$#" -gt 0 ]; then
  COMMAND=$1
  shift
else
  COMMAND=up
fi

AUTH_DATABASE_URL=${AUTH_DATABASE_URL:-postgres://auth_user:auth_pass@auth-postgres:5432/auth_db?sslmode=disable}
EVENT_DATABASE_URL=${EVENT_DATABASE_URL:-postgres://event_user:event_pass@event-postgres:5432/event_db?sslmode=disable}
NOTIFICATION_DATABASE_URL=${NOTIFICATION_DATABASE_URL:-postgres://notification_user:notification_pass@notification-postgres:5432/notification_db?sslmode=disable}

run_migration() {
  service=$1
  database_url=$2
  shift 2

  printf 'running %s migrations for %s\n' "$COMMAND" "$service"
  docker run --rm \
    --network sports-platform \
    -v "$ROOT_DIR/services/$service/migrations:/migrations:ro" \
    "$MIGRATE_IMAGE" \
    -path=/migrations \
    -database "$database_url" \
    "$COMMAND" "$@"
}

run_migration auth-service "$AUTH_DATABASE_URL" "$@"
run_migration event-service "$EVENT_DATABASE_URL" "$@"
run_migration notification-service "$NOTIFICATION_DATABASE_URL" "$@"
