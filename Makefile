COMPOSE := docker compose -f deploy/docker-compose.yml

.PHONY: up down reset migrate run dev build test lint psql redis-cli

up: ## start postgres and redis
	$(COMPOSE) up -d --wait

down: ## stop them, keep data
	$(COMPOSE) down

reset: ## stop them and delete all data
	$(COMPOSE) down -v

migrate: ## apply database migrations
	go run ./cmd/migrate

run:
	go run ./cmd/server

dev: ## hot reload with air
	air

build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

test:
	go test -race ./...

lint:
	go vet ./...

psql:
	psql "postgres://gpu:gpu@localhost:5433/gpu?sslmode=disable"

redis-cli:
	$(COMPOSE) exec redis redis-cli
