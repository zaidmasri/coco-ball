DB_PATH ?= ./northbasis.db
PORT ?= :8080
BIN_DIR := bin
BIN := $(BIN_DIR)/northbasis-cli
IMAGE := northbasis-cli

.PHONY: build serve migrate reset test clean \
	docker-build up down logs \
	dev dev-down dev-logs

## Local Go binary

build:
	go build -o $(BIN) ./cmd/cli

serve:
	go run ./cmd/cli serve --db $(DB_PATH) --port $(PORT)

migrate:
	go run ./cmd/cli migrate --db $(DB_PATH)

reset:
	go run ./cmd/cli reset --db $(DB_PATH)

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR) tmp

## sqlc / sql-migrate

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

migrate-up:
	go run github.com/rubenv/sql-migrate/sql-migrate@latest up -config=dbconfig.yml

migrate-down:
	go run github.com/rubenv/sql-migrate/sql-migrate@latest down -config=dbconfig.yml

## Production image (Coolify deploys via docker-compose.yml)

docker-build:
	docker build -t $(IMAGE) .

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

## Development (hot reload via air, source bind-mounted)

dev:
	docker compose -f docker-compose.dev.yml up --build

dev-down:
	docker compose -f docker-compose.dev.yml down

dev-logs:
	docker compose -f docker-compose.dev.yml logs -f
