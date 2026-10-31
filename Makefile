-include .env

.DEFAULT_GOAL:= help

BINARY_NAME=bin/app.exe
MAIN_FILE=cmd/uptime/main.go

.PHONY: build run clean fmt lint test test_integration

help:
	@echo "run    - run the application"
	@echo "build  - build into bin/"
	@echo "fmt    - format Go code"
	@echo "test   - run tests with coverage"
	@echo "clean  - remove build output"

build:
	go build -o $(BINARY_NAME) $(MAIN_FILE)
run: build
	./$(BINARY_NAME)
lint:
	golangci-lint run
fmt:
	go fmt
test:
	go test -cover ./...
clean: 
	rm -rf bin/

migrate-up:
	goose -dir internal/infra/postgres/migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up
migrate-down:
	goose -dir internal/infra/postgres/migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" down

migrate-test-up:
	goose -dir internal/infra/postgres/migrations postgres \
	"postgres://$$POSTGRES_TEST_USER:$$POSTGRES_TEST_PASSWORD@$$POSTGRES_TEST_HOST:$$POSTGRES_TEST_PORT/$$POSTGRES_TEST_DB?sslmode=disable" up

test-integration: migrate-test-up
	go test -race -v ./internal/infra/postgres/repository -count=1

test-integration-local:
	docker compose --profile test up -d --wait postgres-test
	set -a; . ./.env && set +a && $(MAKE) test-integration
