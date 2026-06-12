.PHONY: all build test clean dev up down lint help

GO := go
DOCKER := docker

SERVICES := api-gateway sfu chat-service room-service presence-service
PKG_DIR := ./pkgs/...

all: test build

build: $(SERVICES)

$(SERVICES):
	@echo "Building $@..."
	cd services/$@ && $(GO) build ./cmd/server/

test:
	@echo "Running all tests..."
	cd pkgs && $(GO) test $(PKG_DIR) -short -count=1
	for svc in $(SERVICES); do \
		cd services/$$svc && $(GO) test ./... -short -count=1; \
	done

test-integration:
	@echo "Running integration tests (requires RabbitMQ + PostgreSQL)..."
	cd pkgs && $(GO) test $(PKG_DIR) -tags=integration -count=1

lint:
	@echo "Running go vet..."
	cd pkgs && $(GO) vet ./...
	for svc in $(SERVICES); do \
		cd services/$$svc && $(GO) vet ./...; \
	done

clean:
	@echo "Cleaning build artifacts..."
	for svc in $(SERVICES); do \
		rm -f services/$$svc/server; \
	done

dev:
	$(DOCKER) compose -f docker-compose.yml -f docker-compose.dev.yml up --build

up:
	$(DOCKER) compose up --build -d

down:
	$(DOCKER) compose down

logs:
	$(DOCKER) compose logs -f

help:
	@echo "VoxMeet — Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build all Go services"
	@echo "  test         Run all unit tests"
	@echo "  test-integration  Run integration tests"
	@echo "  lint         Run go vet on all packages"
	@echo "  clean        Remove build artifacts"
	@echo "  dev          Start dev environment with hot reload"
	@echo "  up           Start all services in background"
	@echo "  down         Stop all services"
	@echo "  logs         Follow service logs"
