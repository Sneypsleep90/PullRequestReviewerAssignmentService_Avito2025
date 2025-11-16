.PHONY: build run test test-integration test-e2e test-all clean docker-build docker-run docker-stop docker-clean migrate-up migrate-down

# Variables
APP_NAME := reviewer-service
DOCKER_IMAGE := $(APP_NAME):latest
DOCKER_CONTAINER := reviewer-app

# Build the application
build:
	go build -o bin/$(APP_NAME) ./cmd/main.go

# Run the application locally
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v ./...

# Run integration tests
test-integration:
	go test -v ./tests/integration/...

# Run E2E tests
test-e2e:
	go test -v ./tests/e2e/...

# Run all tests (integration + e2e)
test-all: test-integration test-e2e

# Clean build artifacts
clean:
	rm -rf bin/

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-clean:
	docker-compose down -v
	docker system prune -f

# Database migrations
migrate-up:
	migrate -path ./migrations -database "postgres://myuser:mypassword@localhost:5432/mydatabase?sslmode=disable" up

migrate-down:
	migrate -path ./migrations -database "postgres://myuser:mypassword@localhost:5432/mydatabase?sslmode=disable" down

# Development workflow
dev: docker-build docker-run

# Production build
prod-build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) ./cmd/main.go

# Lint and format
lint:
	golangci-lint run

fmt:
	go fmt ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Generate OpenAPI
generate-api:
	swag init -g cmd/main.go

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application locally"
	@echo "  test           - Run all tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run E2E tests"
	@echo "  test-all       - Run integration + E2E tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker containers"
	@echo "  docker-clean   - Clean Docker resources"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback database migrations"
	@echo "  dev            - Build and run with Docker"
	@echo "  prod-build     - Build for production"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  deps           - Install and tidy dependencies"
	@echo "  generate-api   - Generate OpenAPI documentation"
	@echo "  help           - Show this help message"
