DATABASE_URL ?= postgres://gpu:gpu@localhost:5433/gpu?sslmode=disable
HTTP_ADDR ?= :8080
export DATABASE_URL HTTP_ADDR

COMPOSE := docker compose -f deploy/docker-compose.yml

.PHONY: up down reset run build test psql

up:
	$(COMPOSE) up -d --wait

down:
	$(COMPOSE) down

reset:
	$(COMPOSE) down -v

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

psql:
	psql "$(DATABASE_URL)"
