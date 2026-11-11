#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if [ -f .env ]; then echo '.env already exists; preserving current keys.'; exit 0; fi
umask 077
TASK_CONFIG=$(mktemp)
trap 'rm -f "$TASK_CONFIG"' EXIT
(cd backend && go run ./cmd/keygen) > "$TASK_CONFIG"
mv "$TASK_CONFIG" .env
echo 'Created .env with local development keys. Do not commit it.'
