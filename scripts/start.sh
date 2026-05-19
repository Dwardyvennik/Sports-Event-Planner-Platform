#!/bin/sh
set -eu

docker compose up --build -d
printf '%s\n' "Platform started. Database migrations run through docker-compose."
