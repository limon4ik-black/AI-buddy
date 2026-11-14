.DEFAULT_GOAL := up
COMPOSE = docker compose --env-file .env -f deploy/compose.yaml
.PHONY: configure up down logs build config rag
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
