.PHONY: proto clean build run test test-coverage test-race test-all test-internal coverage-report docker-build docker-up docker-down

# Proto generation
proto:
	@echo "Generating protobuf files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/iam/v1/*.proto

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	find api/proto -name "*.pb.go" -delete
	rm -f iam-server
	rm -f coverage.out coverage.html

# Build server
build: proto
	@echo "Building server..."
	go build -o iam-server cmd/server/main.go

# Run server
run: build
	@echo "Running server..."
	./iam-server

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Run all tests with coverage and race detection (CI mode)
test-all:
	@echo "Running all tests with coverage and race detection..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | grep total

# Run only internal package tests
test-internal:
	@echo "Running internal package tests..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./internal/...
	go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-report: test-coverage
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t chassis/iam:latest .

docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping services..."
	docker-compose down

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run
