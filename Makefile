ifneq (,$(wildcard .env))
	include .env
	export
endif

PROFILE ?= k3d
CLUSTER_NAME ?= sagars-cluster
APP_NAME ?= go-api
HELM_RELEASE_NAME ?= go-api
HELM_VALUES ?= values-windows-access.yaml # 

# Include domain targets
include make/vars.mk
include make/bootstrap.mk
include make/runtime.mk
include make/app.mk

.PHONY: help setup-tools runtime-up runtime-down build run deploy \
        deploy undeploy helm-lint helm-validate

help:
	@echo "Available targets:"
	@echo ""
	@echo "  Setup & Environment:"
	@echo "    make setup-tools         Install required tools (specify PROFILE)"
	@echo "                             Usage: make setup-tools PROFILE=k3d|aks|common"
	@echo "                             Default profile: k3d"
	@echo ""
	@echo "  Cluster Management:"
	@echo "    make runtime-up          Create and setup a runtime environment (k3d cluster)"
	@echo "    make runtime-down        Shutdown runtime environment (k3d cluster)"
	@echo ""
	@echo "  Go API - Build & Run:"
	@echo "    make build            Build the Go API application"
	@echo "    make run              Run the Go API application locally"
	@echo "    make docker-build     Build Go API Docker image"
	@echo "    make docker-import    Import Go API Docker image into k3d"
	@echo ""
	@echo "  Go API - Kubernetes Deployment:"
	@echo "    make deploy       Deploy $(APP_NAME) to k3d via Helm"
	@echo "                             Usage: make deploy HELM_VALUES=values-dev.yaml"
	@echo "                             Default values: values-windows-access.yaml"
	@echo "    make undeploy     Uninstall $(APP_NAME) from k3d"
	@echo "    make helm-lint           Validate Helm chart syntax"
	@echo "    make helm-validate       Preview Helm manifests (dry-run)"

