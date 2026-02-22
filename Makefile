ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: help setup-tools cluster-up cluster-down go-build go-run go-docker-build

help:
	@echo "Available targets:"
	@echo "  make setup-tools           Install required tools (specify PROFILE)"
	@echo "                             Usage: make setup-tools PROFILE=k3d|aks|common"
	@echo "                             Default profile: k3d"
	@echo "  make cluster-up            Create and setup a k3d cluster"
	@echo "  make cluster-down          Shutdown k3d cluster"
	@echo "  make go-build              Build the Go API application"
	@echo "  make go-run                Run the Go API application locally"
	@echo "  make go-docker-build       Build Go API Docker image"

setup-tools:
	@echo "Installing required tools for profile: $(PROFILE)..."
	@bash bootstrap/setup-tools.sh $(PROFILE)

cluster-up:
	@echo "Creating k3d cluster..."
	@bash runtimes/k3d/up.sh $(CLUSTER_NAME)

cluster-down:
	@echo "Shutting down k3d cluster..."
	@bash runtimes/k3d/down.sh

go-build:
	@echo "Building Go API..."
	@cd apps/go-api && go mod tidy && go build -o app .

go-run:
	@echo "Running Go API..."
	@cd apps/go-api && go run main.go

go-docker-build:
	@echo "Building Go API Docker image..."
	@docker build -t go-api:latest apps/go-api/