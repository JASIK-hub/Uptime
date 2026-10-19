.DEFAULT_GOAL:= help

BINARY_NAME=bin/app.exe
MAIN_FILE=cmd/uptime/main.go

.PHONY: build run clean fmt lint test

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