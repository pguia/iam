.PHONY: proto clean build run test docker-build docker-up docker-down

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

# Build server
build: proto
	@echo "Building server..."
	go build -o iam-server cmd/server/main.go

# Run server
run: build
	@echo "Running server..."
	./iam-server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

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
