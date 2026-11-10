#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
bash scripts/test-backend.sh
(cd frontend && npm ci --no-audit --no-fund && npm test && npm run build)
python3 scripts/coverage.py
python3 scripts/check-contracts.py
