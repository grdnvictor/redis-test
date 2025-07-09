# Redis-Go Makefile

.PHONY: run build test clean dev deps docker-build docker-run docker-test

# Variables
BINARY_NAME=redis-go
BUILD_DIR=bin
MAIN_FILE=main.go

# Commandes principales
run:
	@echo "Starting Redis-Go server..."
	go run $(MAIN_FILE)

build:
	@echo "Building Redis-Go..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Binary created at $(BUILD_DIR)/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	go clean

# Développement
dev:
	@echo "Starting in development mode with hot reload..."
	go run $(MAIN_FILE)

deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod verify

# Tests avancés
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

race:
	@echo "Running race condition tests..."
	go test -race ./...

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t redis-go .

docker-run: docker-build
	@echo "Running Redis-Go in Docker..."
	docker run -p 6379:6379 redis-go

docker-compose-up:
	@echo "Starting full environment with docker-compose..."
	docker-compose up --build

docker-compose-test:
	@echo "Running tests with docker-compose..."
	docker-compose up --build redis-test

docker-cli:
	@echo "Starting interactive redis-cli..."
	docker-compose up --build redis-cli

# Linting et formatage
fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

# Installation globale
install: build
	@echo "Installing Redis-Go globally..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Tests avec le vrai redis-cli
test-with-redis:
	@echo "Testing with real redis-cli..."
	@chmod +x test-scripts/comprehensive-test.sh
	@./test-scripts/comprehensive-test.sh localhost 6379

# Help
help:
	@echo "Available commands:"
	@echo "  run                - Start the Redis server"
	@echo "  build              - Build the binary"
	@echo "  test               - Run Go tests"
	@echo "  clean              - Clean build artifacts"
	@echo "  dev                - Start in development mode"
	@echo "  deps               - Install/update dependencies"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  install            - Install globally"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run in Docker"
	@echo "  docker-compose-up  - Start full environment"
	@echo "  docker-compose-test- Run automated tests"
	@echo "  docker-cli         - Interactive redis-cli"
	@echo "  test-with-redis    - Test with real redis-cli"