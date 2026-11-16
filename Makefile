.DEFAULT_GOAL := up
export COMPOSE_PARALLEL_LIMIT ?= 2
# Docker bridge DNS can be unreachable behind a Linux VPN.
# Override with BUILD_NETWORK=default to use Docker's regular build network.
BUILD_NETWORK ?= $(if $(filter Linux,$(shell uname -s)),host,default)
export BUILD_NETWORK
COMPOSE = docker compose --env-file .env -f deploy/compose.yaml -f deploy/build-host.yaml
METRICS = docker compose --env-file .env -f observability/compose.yaml
.PHONY: configure up down logs build config rag rag-import metrics-up metrics-down metrics-logs
configure:
	@sh deploy/configure.sh
up: configure
	$(COMPOSE) up -d --build --wait --wait-timeout 1200
down:
	$(METRICS) down --remove-orphans 2>/dev/null || true
	$(COMPOSE) --profile rag down --remove-orphans
logs:
	$(COMPOSE) logs -f
metrics-up: configure
	$(METRICS) up -d
metrics-down:
	$(METRICS) down --remove-orphans 2>/dev/null || true
metrics-logs:
	$(METRICS) logs -f
build: configure
	$(COMPOSE) build
config: configure
	$(COMPOSE) config --quiet
rag: configure
	$(COMPOSE) -f deploy/rag.yaml --profile rag up -d --build --wait --wait-timeout 1800
rag-import: configure
	$(COMPOSE) --profile rag up -d --wait rag_postgres
	$(COMPOSE) --profile rag build rag
	$(COMPOSE) --profile rag run --rm --no-deps rag python -m src.knowledge import
