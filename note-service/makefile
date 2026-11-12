include .env
export

export PROJECT_ROOT=$(shell pwd)

# Цели-команды: имена совпадают с каталогами проекта (docs/), поэтому make нужно
# сказать, что это не файлы — иначе он решит, что цель «уже собрана».
.PHONY: run test_jwt_token docs env-up env-down env-cleanup migrate-create migrate-up migrate-down migrate-action env-port-forward env-port-close service-deploy service-down

run:
	go run ./cmd/note-service

# Локальный тестовый токен. Приватный seed читается из .env репы по месту,
# в .env сервиса его держать не нужно.
test_jwt_token:
	@if [ -z "$(user_id)" ]; then \
		echo "Отсутствует параметр user_id"; \
		exit 1; \
	fi; \
	seed=$$(grep '^JWT_PRIVATE_SEED=' ../.env 2>/dev/null | cut -d= -f2-); \
	if [ -z "$$seed" ]; then \
		echo "Не нашёл JWT_PRIVATE_SEED в ../.env"; \
		exit 1; \
	fi; \
	JWT_PRIVATE_SEED="$$seed" go run ./cmd/test_jwt_token -user_id "$(user_id)"

# Генерирует docs/ (openapi) из аннотаций в коде. swag крутится в контейнере, локально не устанавливается.
docs:
	docker compose run --rm swag

# Поднимает БД и собирает/запускает сервис в докере. Сервис ходит в БД по сети compose,
# порт 5432 наружу ему не нужен (форвардер остаётся для локального make run).
service-deploy:
	docker compose up -d --build note_postgres note_service

# Обратная операция: убирает контейнеры сервиса и БД. Данные остаются в out/.pgdata;
# снести вместе с ними — make env-cleanup.
service-down:
	docker compose down note_service note_postgres

env-up:
	docker compose up -d note_postgres

env-down:
	docker compose down note_postgres

env-cleanup:
	@read -p "!Произойдет полная отчистка директории out и остановка контейнеров. Продолжить? [Y/n]: " ans; \
	if [ "$$ans" = "Y" ]; then \
		docker compose down note_postgres port_forwarder && \
		sudo rm -rf ${PROJECT_ROOT}/out/.pgdata && \
		echo "Очистка окружения выполненна"; \
	else \
		echo "Очистка окружения отменена"; \
	fi


migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует параметр seq"; \
		exit 1; \
	fi; \
	docker compose run --rm note_postgres_migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@$(MAKE) --no-print-directory migrate-action action=up

migrate-down:
	@$(MAKE) --no-print-directory migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует параметр action"; \
		exit 1; \
	fi; \
	docker compose run --rm note_postgres_migrate \
		-path /migrations \
		-database postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@note_postgres:5432/$(POSTGRES_DB)?sslmode=disable \
		"$(action)"

env-port-forward:
	@docker compose up -d port_forwarder

env-port-close:
	@docker compose down port_forwarder
