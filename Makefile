ifneq (,$(wildcard .env))
include .env
export
endif

OPENAPI_SPEC   := contracts/openapi/trip-service.openapi.yaml
OAPI_CONFIG    := internal/generated/oapi-codegen.yaml
MIGRATIONS_DIR := migrations
BIN            := bin/trip-service

GOOSE := go tool goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)"

.PHONY: generate migrate migrate-down migrate-reset migrate-status build run test test-integration lint

generate:
	go tool oapi-codegen -config $(OAPI_CONFIG) $(OPENAPI_SPEC)
	go generate ./internal/...

migrate:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-reset:
	$(GOOSE) reset

migrate-status:
	$(GOOSE) status

build:
	go build -o $(BIN) ./cmd/trip-service

run: build
	./$(BIN)

test:
	go test -race -cover ./...

test-integration:
	go test -race -tags=integration -count=1 ./internal/repository/trip

lint:
	golangci-lint run ./...
