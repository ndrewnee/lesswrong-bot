# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Telegram bot written in Go that aggregates posts from rationalist blogs (LessWrong.ru, Slate Star Codex, Astral Codex Ten, LessWrong.com). The bot allows users to read random posts or view top posts from these sources.

## Common Commands

### Development
- `make run` - Run the application locally (requires .env file)
- `make test` - Run unit tests with race detection
- `make test-integration` - Run integration tests (requires .env.test file)
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run golangci-lint
- `make build` - Build the binary
- `make clean` - Clean build artifacts

### Dependencies
- `make go-mod-tidy` - Tidy go modules
- `make go-mod-download` - Download go modules

### Docker
- `docker-compose up` - Run bot with Redis in containers
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container

## Architecture

### Core Components

**Main Entry Point** (`main.go`): Initializes storage (Redis with memory fallback), creates bot instance, and starts message processing loop.

**Bot Package** (`bot/`): Contains core bot logic with dependency injection pattern:
- `bot.go` - Main bot struct with Telegram API integration
- `random.go` - Random post fetching logic  
- `top.go` - Top posts functionality
- `source.go` - Source switching between different blog sites
- `http.go` - HTTP client wrapper
- `utils.go` - Utility functions

**Models** (`models/`): Data structures for posts and sources.

**Storage** (`storage/`): Pluggable storage interface with Redis and in-memory implementations.

**Config** (`config/`): Environment-based configuration parsing.

### Key Patterns

The codebase uses dependency injection with the `Options` pattern for testability. The bot supports both polling and webhook modes for Telegram updates. Storage is abstracted behind an interface allowing fallback from Redis to memory cache.

### Testing

The project includes both unit tests and integration tests. Integration tests use the `-tags=integration` build flag and require a separate `.env.test` file. Test data is stored in `bot/testdata/` with JSON responses and expected markdown outputs.

### Environment Variables

Key variables in `.env`:
- `TOKEN` - Telegram bot token (required)
- `REDIS_URL` - Redis connection string (optional, falls back to memory)
- `WEBHOOK` - Enable webhook mode vs polling
- `DEBUG` - Enable debug logging

Integration tests require `.env.test` with test-specific values.