ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: help setup-tools cluster-up

help:
	@echo "Available targets:"
	@echo "  make setup-tools    Install required tools (docker, kubectl, k3d)"
	@echo "  make cluster-up     Create and setup a k3d cluster"

setup-tools:
	@echo "Installing required tools..."
	@bash scripts/setup-tools.sh

cluster-up:
	@echo "Creating k3d cluster..."
	@bash scripts/cluster-up.sh $(CLUSTER_NAME)

cluster-down:
	@echo "Shutting down k3d cluster..."
	@bash scripts/cluster-down.sh
