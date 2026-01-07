.PHONY: test test-unit test-integration test-coverage test-all build run

# Run all tests
test-all: test-unit test-integration

# Run unit tests
test-unit:
	go test -v -short ./...

# Run integration tests
test-integration:
	go test -v -tags=integration ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests for specific package
test-pkg:
	go test -v ./$(PKG)

# Build the application
build:
	go build -o bin/api ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Install test dependencies
test-deps:
	go get -u github.com/stretchr/testify/assert
	go get -u github.com/stretchr/testify/mock
	go get -u github.com/stretchr/testify/require

# Clean test artifacts
clean:
	rm -f coverage.out coverage.html
	rm -rf bin/

