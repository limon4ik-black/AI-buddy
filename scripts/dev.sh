#!/usr/bin/env bash
# Local alternative to Docker: isolated PostgreSQL + six Go processes + Vite.
set -euo pipefail
cd "$(dirname "$0")/.."
bash scripts/configure.sh
set -a
source .env
set +a
TASK_DEV_DIR="$PWD/.local"
mkdir -p "$TASK_DEV_DIR/bin" "$TASK_DEV_DIR/log"
TASK_PIDS=''
cleanup() {
 for pid in $TASK_PIDS; do kill "$pid" 2>/dev/null || true; done
 if [ "${TASK_PG_STARTED:-false}" = true ]; then pg_ctl -D "$TASK_DEV_DIR/pg" -m fast -w stop >/dev/null; fi
}
trap cleanup EXIT INT TERM
export PGHOST=127.0.0.1 PGPORT=${DEV_PGPORT:-55432} PGUSER=buddy_admin
if [ ! -f "$TASK_DEV_DIR/pg/PG_VERSION" ]; then
 initdb -D "$TASK_DEV_DIR/pg" -U buddy_admin -A trust --no-locale -E UTF8 >/dev/null
 pg_ctl -D "$TASK_DEV_DIR/pg" -l "$TASK_DEV_DIR/log/postgres.log" -o "-p $PGPORT -h $PGHOST -k ${TMPDIR:-/tmp}" -w start >/dev/null
 TASK_PG_STARTED=true

elif ! pg_ctl -D "$TASK_DEV_DIR/pg" status >/dev/null 2>&1; then
 pg_ctl -D "$TASK_DEV_DIR/pg" -l "$TASK_DEV_DIR/log/postgres.log" -o "-p $PGPORT -h $PGHOST -k ${TMPDIR:-/tmp}" -w start >/dev/null
 TASK_PG_STARTED=true
fi
if [ ! -f "$TASK_DEV_DIR/pg/.buddy-initialized" ]; then
 POSTGRES_USER=buddy_admin bash infra/init-db.sh >/dev/null
 touch "$TASK_DEV_DIR/pg/.buddy-initialized"
fi
for name in auth info chat gateway ragmock corpmock; do (cd backend && go build -o "$TASK_DEV_DIR/bin/$name" "./cmd/$name"); done
export APP_ORIGIN=http://localhost:5173 AUTH_URL=http://127.0.0.1:8081 INFO_URL=http://127.0.0.1:8082 CHAT_URL=http://127.0.0.1:8083 CORP_URL=http://127.0.0.1:8085
export RAG_URL=${RAG_URL:-http://127.0.0.1:8084}
start() {
 local name=$1 port=$2 db=${3:-} db_user=${4:-} db_password=${5:-}
 (export LISTEN_ADDR="127.0.0.1:$port";if [ -n "$db" ]; then export DATABASE_URL="postgres://$db_user:$db_password@127.0.0.1:$PGPORT/$db?sslmode=disable";fi
 if [ "$name" != auth ]; then unset JWT_PRIVATE_SEED; fi
 exec "$TASK_DEV_DIR/bin/$name") > "$TASK_DEV_DIR/log/$name.log" 2>&1 &
 TASK_PIDS="$TASK_PIDS $!"
}
start auth 8081 auth_db auth_user "$AUTH_DB_PASSWORD"
start info 8082 info_db info_user "$INFO_DB_PASSWORD"
start chat 8083 chat_db chat_user "$CHAT_DB_PASSWORD"
start ragmock 8084
start corpmock 8085
start gateway 8080
if [ ! -d frontend/node_modules ]; then (cd frontend && npm ci --no-audit --no-fund); fi
(cd frontend && exec npm run dev -- --host 127.0.0.1) > "$TASK_DEV_DIR/log/frontend.log" 2>&1 &
TASK_PIDS="$TASK_PIDS $!"
echo 'AI Buddy: http://localhost:5173 — Info API: http://localhost:8082'
echo 'Test login: intern@example.test / DemoPassword123!'
echo 'Logs: .local/log — press Ctrl+C to stop.'
wait
