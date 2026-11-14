#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if [ -f .env ]; then
  echo '.env exists; preserving configuration.'
  exit 0
fi
umask 077
config_tmp=$(mktemp)
trap 'rm -f "$config_tmp"' EXIT
docker run --rm --network none -v "$PWD/deploy/keygen/main.go:/src/main.go:ro" -w /src golang:1.25-alpine go run main.go > "$config_tmp"
mv "$config_tmp" .env
echo 'Created .env. Local email inbox: http://localhost:8025'
