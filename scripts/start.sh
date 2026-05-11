#!/bin/sh
set -eu

docker compose up --build -d
printf '%s\n' "Platform started. Run './scripts/migrate.sh up' after databases are healthy."
