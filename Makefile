.DEFAULT_GOAL := up
.PHONY: configure up info test down logs build fmt vet test-backend
configure:
	@if [ ! -f .env ]; then \
		set -eu; \
		umask 077; \
		task_config=$$(mktemp); \
		trap 'rm -f "$$task_config"' EXIT; \
		docker run --rm --network none \
			--mount "type=bind,source=$(CURDIR)/backend/cmd/keygen/main.go,target=/src/main.go,readonly" \
			--workdir /src golang:1.24.13-alpine go run main.go > "$$task_config"; \
		mv "$$task_config" .env; \
	fi
up: configure
	docker compose up --build -d
info: configure
	docker compose up --build -d info
test:
	bash scripts/test.sh
down:
	docker compose down
logs:
	docker compose logs -f

build:
	go -C backend build ./cmd/...
fmt:
	gofmt -w backend/cmd backend/internal backend/migrations backend/tests
vet:
	go -C backend vet ./...
test-backend:
	bash scripts/test-backend.sh

note-service-run:
	@go run backend/note-service/cmd/note-service/main.go