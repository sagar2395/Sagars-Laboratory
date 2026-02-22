ifneq (,$(wildcard .env))
	include .env
	export
endif

PROFILE ?= k3d
CLUSTER_NAME ?= sagar-lab
HELM_RELEASE_NAME ?= go-api
HELM_VALUES ?= values.yaml

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
	@echo "    make cluster-up          Create and setup a k3d cluster"
	@echo "    make cluster-down        Shutdown k3d cluster"
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

go-docker-import:
	@echo "Importing Go API Docker image into k3d cluster..."
	@k3d image import go-api:latest -c $(CLUSTER_NAME)
	@echo "Image imported successfully"

deploy-go-api:
	@echo "Deploying Go API to k3d cluster..."
	@helm lint apps/go-api/deploy/helm/go-api > /dev/null || exit 1
	@# Use helm upgrade --install to handle both fresh installs and updates gracefully
	@helm upgrade --install $(HELM_RELEASE_NAME) apps/go-api/deploy/helm/go-api \
		-f apps/go-api/deploy/helm/go-api/$(HELM_VALUES) \
		--namespace go-api --create-namespace
	@echo ""
	@echo "Deployment complete! Access the application:"
	@echo "  - HTTP: http://go-api.k3d.local"
	@echo "  - Metrics: http://go-api.k3d.local/metrics"
	@echo ""
	@echo "View deployment status:"
	@echo "  kubectl get deployments -n go-api"
	@echo "  kubectl get pods -n go-api"
	@echo "  kubectl get svc -n go-api"

undeploy-go-api:
	@echo "Uninstalling Go API from k3d cluster..."
	@helm uninstall $(HELM_RELEASE_NAME) -n go-api
	@echo "Uninstall complete"

helm-lint:
	@echo "Linting Helm chart..."
	@helm lint apps/go-api/deploy/helm/go-api -f apps/go-api/deploy/helm/go-api/$(HELM_VALUES)
	@echo "Lint complete"

helm-validate:
	@echo "Validating Helm chart (dry-run)..."
	@echo "Using values file: apps/go-api/deploy/helm/go-api/$(HELM_VALUES)"
	@helm template $(HELM_RELEASE_NAME) apps/go-api/deploy/helm/go-api \
		-f apps/go-api/deploy/helm/go-api/$(HELM_VALUES) \
		--namespace go-api
