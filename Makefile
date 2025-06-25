.PHONY: run test test-integration lint build deploy clean docker-build docker-run help

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=lesswrong-bot
DOCKER_IMAGE=lesswrong-bot

run: ## Run the application
	go run .

test: ## Run tests with race detection
	go test -v -race ./...

test-integration: ## Run integration tests
	go test -v -race -tags=integration ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

build: ## Build the application
	go build -o $(BINARY_NAME) .

build-optimized: ## Build with optimizations
	go build -ldflags="-w -s" -o $(BINARY_NAME) .

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

deploy: ## Deploy to Heroku
	git push heroku main

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run -p 9999:9999 $(DOCKER_IMAGE)

go-mod-tidy: ## Tidy go modules
	go mod tidy

go-mod-verify: ## Verify go modules
	go mod verify

go-mod-download: ## Download go modules
	go mod download

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
