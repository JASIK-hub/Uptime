.DEFAULT_GOAL:= help

.PHONY: build run clean fmt lint test

help:
	@echo "run    - run the application"
	@echo "build  - build into bin/"
	@echo "fmt    - format Go code"
	@echo "test   - run tests with coverage"
	@echo "clean  - remove build output"

build:
	go build -o bin/app.exe cmd/uptime/main.go
run:
	go run cmd/uptime
lint:
	golangci-lint run
fmt:
	go fmt
test:
	go test -cover ./...
clean: 
	rm -rf bin/