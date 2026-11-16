#!/bin/sh
# Isolated tests: no application .env, ports, running services or app volumes.
set -eu
cd "$(dirname "$0")/.."
test_root=$PWD
test_output=$(mktemp -d)
test_id="ai-buddy-tests-$$"
cleanup() {
  test_status=$?
  trap - EXIT INT TERM
  docker rm -f "$test_id-go" "$test_id-frontend" "$test_id-rag" "$test_id-db" >/dev/null 2>&1 || true
  docker network rm "$test_id" >/dev/null 2>&1 || true
  if [ "$test_status" -eq 0 ]; then
    rm -rf "$test_output"
  else
    echo "Test logs retained outside repository: $test_output" >&2
  fi
  exit "$test_status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
docker info >/dev/null
docker network create "$test_id" >/dev/null
docker run -d --name "$test_id-db" --network "$test_id" --network-alias database \
  -e POSTGRES_HOST_AUTH_METHOD=trust --tmpfs /var/lib/postgresql/data \
  pgvector/pgvector:pg17 >/dev/null
test_ready=false
for attempt in $(seq 1 30); do
  if docker exec "$test_id-db" pg_isready -U postgres >/dev/null 2>&1; then test_ready=true; break; fi
  sleep 1
done
[ "$test_ready" = true ] || { echo 'Test database did not start' >&2; exit 1; }

echo 'Testing Go: root, chat-service, note-service (real disposable PostgreSQL)'
docker run --rm --name "$test_id-go" --cpus=2 --network "$test_id" \
  -e GOMAXPROCS=2 -e TEST_DB_ADMIN_URL='postgres://postgres@database/postgres?sslmode=disable' \
  -e CHAT_TEST_DATABASE_URL='postgres://postgres@database/postgres?sslmode=disable' \
  -v "$test_root:/src:ro" -v "$test_output:/reports" \
  -v buddy-coverage-gomod:/go/pkg/mod -v buddy-coverage-gobuild:/root/.cache/go-build \
  -w /src golang:1.27-bookworm sh -ec '
    go test -json -p 2 -count=1 -covermode=atomic \
      -coverpkg=./...,github.com/limon4ik-black/AI-buddy/chat-service/internal/...,github.com/limon4ik-black/AI-buddy/note-service/onboarding \
      -coverprofile=/reports/root.out ./... > /reports/root-tests.json
    cd /src/chat-service
    go test -json -p 2 -count=1 -covermode=atomic -coverpkg=./... -coverprofile=/reports/chat.out ./... > /reports/chat-tests.json
    cd /src/note-service
    go test -json -p 2 -count=1 -covermode=atomic -coverpkg=./... -coverprofile=/reports/note.out ./... > /reports/note-tests.json
  '

echo 'Testing frontend: Vitest/V8 (including BoardEditor wrapper)'
docker run --rm --name "$test_id-frontend" --cpus=2 \
  -v "$test_root/frontend:/src:ro" -v buddy-coverage-node:/src/node_modules \
  -v "$test_output:/reports" -w /src node:22-alpine sh -ec '
    npm ci --no-audit --no-fund > /reports/npm-install.txt 2>&1
    npx vitest run --coverage --coverage.reportsDirectory=/reports/frontend \
      --coverage.reporter=json-summary --coverage.reporter=json --coverage.reporter=text \
      > /reports/frontend-tests.txt 2>&1
  '

echo 'Testing RAG: pytest/coverage.py + real pgvector; LLM and embeddings mocked'
docker build -q -t ai-buddy-coverage-rag "$test_root/rag-service" > "$test_output/python-image.txt"
docker run --rm --name "$test_id-rag" --cpus=2 --network "$test_id" \
  -e RAG_TEST_DATABASE_URL='postgres://postgres@database/postgres?sslmode=disable' \
  -e COVERAGE_FILE=/reports/.coverage -e OMP_NUM_THREADS=2 \
  -v "$test_root/rag-service:/work:ro" -v "$test_output:/reports" -w /work \
  --entrypoint sh ai-buddy-coverage-rag -ec '
    uv pip install --python /app/.venv/bin/python pytest==9.2.0 pytest-cov==7.0.0 coverage==7.17.3 > /reports/python-install.txt 2>&1
    /app/.venv/bin/python -m pytest -q -p no:cacheprovider --cov=src \
      --cov-report=json:/reports/python.json --cov-report=term-missing > /reports/python-tests.txt 2>&1
  '

echo 'Checking executable-line coverage and saving compact reports'
docker run --rm --network none -v "$test_root:/src" -v "$test_output:/raw:ro" \
  -w /src python:3.14-slim python deploy/coverage.py /raw
