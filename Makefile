ifneq (,$(wildcard .env))
	include .env
	export
endif

PROFILE ?= k3d
CLUSTER_NAME ?= sagars-cluster
APP_NAME ?= go-api
HELM_RELEASE_NAME ?= go-api
HELM_VALUES ?= values.yaml

# Include domain targets
include make/vars.mk
include make/bootstrap.mk
include make/runtime.mk
include make/app.mk

.PHONY: help setup-tools cluster-up cluster-down go-build go-run go-docker-build \
        go-docker-import deploy-go-api undeploy-go-api helm-lint helm-validate

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
	@echo "    make go-build            Build the Go API application"
	@echo "    make go-run              Run the Go API application locally"
	@echo "    make go-docker-build     Build Go API Docker image"
	@echo "    make go-docker-import    Import Go API Docker image into k3d"
	@echo ""
	@echo "  Go API - Kubernetes Deployment:"
	@echo "    make deploy-go-api       Deploy Go API to k3d via Helm"
	@echo "                             Usage: make deploy-go-api HELM_VALUES=values-dev.yaml"
	@echo "                             Default values: values.yaml"
	@echo "    make undeploy-go-api     Uninstall Go API from k3d"
	@echo "    make helm-lint           Validate Helm chart syntax"
	@echo "    make helm-validate       Preview Helm manifests (dry-run)"

