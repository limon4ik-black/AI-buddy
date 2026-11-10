#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p reports
TASK_TEST_CLUSTER=''
cleanup() { if [ -n "$TASK_TEST_CLUSTER" ]; then pg_ctl -D "$TASK_TEST_CLUSTER" -m fast -w stop >/dev/null; rm -rf "$TASK_TEST_CLUSTER"; fi; }
trap cleanup EXIT
if [ -z "${TEST_DB_ADMIN_URL:-}" ]; then
 command -v initdb >/dev/null || { echo 'Provide TEST_DB_ADMIN_URL for a disposable PostgreSQL server, or install PostgreSQL locally.'; exit 1; }
 TASK_TEST_CLUSTER=$(mktemp -d "${TMPDIR:-/tmp}/buddy-tests.XXXXXXXX")
 TEST_PGPORT=${TEST_PGPORT:-55439}
 initdb -D "$TASK_TEST_CLUSTER" -U buddy_test -A trust --no-locale -E UTF8 >/dev/null
 pg_ctl -D "$TASK_TEST_CLUSTER" -l "$TASK_TEST_CLUSTER/server.log" -o "-p $TEST_PGPORT -h 127.0.0.1 -k $TASK_TEST_CLUSTER" -w start >/dev/null
 export TEST_DB_ADMIN_URL="postgres://buddy_test@127.0.0.1:$TEST_PGPORT/postgres?sslmode=disable"
fi
export CONTRACT_FIXTURES_OUT="$PWD/reports/contract-fixtures.json"
(cd backend && go test -race -count=1 -coverpkg=./... -coverprofile=../reports/backend.cover ./... && go tool cover -func=../reports/backend.cover > ../reports/backend-statements.txt && go tool cover -html=../reports/backend.cover -o ../reports/backend.html)
