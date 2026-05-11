#!/bin/sh
set -eu

host=$1
port=$2
timeout=${3:-30}

start=$(date +%s)
while :; do
  if nc -z "$host" "$port" >/dev/null 2>&1; then
    exit 0
  fi

  now=$(date +%s)
  if [ $((now - start)) -ge "$timeout" ]; then
    printf 'timeout waiting for %s:%s\n' "$host" "$port" >&2
    exit 1
  fi

  sleep 1
done
