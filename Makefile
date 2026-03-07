ifneq (,$(wildcard .env))
	include .env
	export
endif

PROFILE ?= k3d
CLUSTER_NAME ?= sagars-cluster
APP_NAME ?= go-api
HELM_RELEASE_NAME ?= go-api
HELM_VALUES ?= values-dev.yaml # 

# default values are read from .env if present; individual commands may
# override them by passing VAR=val on the make command line.

# Include domain targets
include make/vars.mk
include make/bootstrap.mk
include make/runtime.mk
include make/app.mk
include make/platform.mk
include make/check.mk
include make/cli.mk
include make/services.mk
include make/terraform.mk

.PHONY: help setup-tools runtime-up runtime-down runtime-status build run deploy \
        destroy-app lint validate deploy-all destroy-all-apps \
        init teardown reset check-tools check-cluster check-ingress platform-up platform-down platform-status \
        terraform-init terraform-plan terraform-apply terraform-destroy terraform-output terraform-status

# lifecycle helpers
init:
	@$(MAKE) setup-tools PROFILE=$(PROFILE)
	@$(MAKE) runtime-up
	@$(MAKE) platform-up

teardown:
	@$(MAKE) destroy-all-apps || true
	@$(MAKE) platform-down || true
	@$(MAKE) runtime-down || true

reset: teardown init

run:
	@$(MAKE) local-run APP_NAME=$(APP_NAME)

help:
	@echo "Available targets:"
	@echo ""
	@echo "  Setup & Environment:"
	@echo "    make setup-tools         Install required CLI tools (PROFILES: k3d, aks, eks, common)"
	@echo "                             Default profile: $(PROFILE)"
	@echo "    make check-tools         Verify required binaries are in PATH"
	@echo ""
	@echo "  Cluster Management (PROFILE=$(PROFILE)):"
	@echo "    make runtime-up          Create cluster (k3d local, AKS via Terraform, EKS via Terraform)"
	@echo "    make runtime-down        Destroy the current cluster"
	@echo "    make runtime-status      Show cluster info and nodes"
	@echo "    make init                setup-tools + runtime-up + platform-up"
	@echo "    make teardown            destroy-apps + platform-down + runtime-down"
	@echo "    make reset               Run teardown then init (recreate from scratch)"
	@echo "    make check-cluster       Ensure current kubecontext can reach the cluster"
	@echo "    make check-ingress       Confirm platform ingress controller is ready"
	@echo ""
	@echo "  Application tasks:"
	@echo "    make build APP_NAME=<name>       Run the build strategy defined in apps/<name>/app.env"
	@echo "    make run  APP_NAME=<name>        Execute the application locally (if supported)"
	@echo "    make deploy APP_NAME=<name>      Run the deploy strategy (helm, lambda, etc.)"
	@echo "    make destroy-app APP_NAME=<name> Remove the deployed application"
	@echo "    make lint  APP_NAME=<name>       Run strategy-specific lint/validation"
	@echo "    make validate APP_NAME=<name>    Preview manifests or perform dry-run"
	@echo ""
	@echo "  Bulk operations:"
	@echo "    make deploy-all       Deploy every app in the apps/ directory"
	@echo "    make destroy-all-apps Uninstall every deployed app"
	@echo ""
	@echo "  CLI (labctl):"
	@echo "    make cli-build        Build the labctl Go binary to bin/labctl"
	@echo "    make cli-install      Build and install labctl to PATH"
	@echo "    make cli-tidy         Run go mod tidy for the CLI module"
	@echo "    make cli-clean        Remove the labctl binary"
	@echo ""
	@echo "  Shared Services:"
	@echo "    make service-list                List available shared services"
	@echo "    make service-up SERVICE=<name>   Install a shared service (e.g. redis)"
	@echo "    make service-down SERVICE=<name> Uninstall a shared service"
	@echo "    make service-status              Show status of all services"
	@echo "    make service-status SERVICE=<name> Show status of a specific service"
	@echo ""
	@echo "  Terraform (cloud runtimes: TF_ENV=dev|staging, PROFILE=aks|eks):"
	@echo "    make terraform-init              Initialize Terraform working directory"
	@echo "    make terraform-plan              Preview infrastructure changes"
	@echo "    make terraform-apply             Apply infrastructure changes"
	@echo "    make terraform-destroy           Destroy cloud infrastructure"
	@echo "    make terraform-output            Show Terraform outputs"
	@echo "    make terraform-status            Show current Terraform state"
	@echo ""
	@echo "Usage notes: configuration for each app lives in apps/<name>/app.env"
	@echo "Variables such as BUILD_STRATEGY, DEPLOY_STRATEGY, HELM_VALUES, etc. are set there."
	@echo "Most variables may also be overridden on the command line:"
	@echo "  make build APP_NAME=foo BUILD_STRATEGY=golang"

