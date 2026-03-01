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

.PHONY: help setup-tools runtime-up runtime-down build run deploy \
        destroy-app lint validate deploy-all destroy-all-apps \
        init teardown reset check-tools check-cluster check-ingress platform-up platform-down platform-status

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
	@echo "    make setup-tools         Install required CLI tools (PROFILES: k3d, aks, common)"
	@echo "                             Default profile: $(PROFILE)"
	@echo "    make check-tools         Verify required binaries are in PATH"
	@echo ""
	@echo "  Cluster Management:"
	@echo "    make runtime-up          Create a runtime (e.g. k3d cluster) based on PROFILE"
	@echo "    make runtime-down        Destroy the current runtime"
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
	@echo "Usage notes: configuration for each app lives in apps/<name>/app.env"
	@echo "Variables such as BUILD_STRATEGY, DEPLOY_STRATEGY, HELM_VALUES, etc. are set there."
	@echo "Most variables may also be overridden on the command line:"
	@echo "  make build APP_NAME=foo BUILD_STRATEGY=golang"

