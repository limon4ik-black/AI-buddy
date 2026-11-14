.DEFAULT_GOAL := up
COMPOSE = docker compose --env-file .env -f deploy/compose.yaml
.PHONY: configure up down logs build config rag rag-import
configure:
	@sh deploy/configure.sh
up: configure
	$(COMPOSE) up -d --build
down:
	$(COMPOSE) --profile rag down --remove-orphans
logs:
	$(COMPOSE) logs -f
build: configure
	$(COMPOSE) build
config: configure
	$(COMPOSE) config --quiet
rag: configure
	$(COMPOSE) -f deploy/rag.yaml --profile rag up -d --build
rag-import: configure
	$(COMPOSE) --profile rag up -d --wait rag_postgres
	$(COMPOSE) --profile rag build rag
	$(COMPOSE) --profile rag run --rm --no-deps rag python -m src.knowledge import
