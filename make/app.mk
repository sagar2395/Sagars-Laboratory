APPS := $(shell ls apps)

local-build:
	@echo "Building Go API..."
	@cd apps/$(APP_NAME) && go mod tidy && go build -o app .

local-run:
	@echo "Running Go API..."
	@cd apps/$(APP_NAME) && go run main.go

docker-build:
	@echo "Building Go API Docker image..."
	@docker build -t $(APP_NAME):latest apps/$(APP_NAME)/
ifeq ($(PROFILE),k3d)
	@echo "Importing Go API Docker image into k3d cluster..."
	@k3d image import $(APP_NAME):latest -c $(CLUSTER_NAME)
	@echo "Image imported successfully"
endif

deploy:
	@echo "Deploying Go API to k3d cluster..."
	@helm lint apps/$(APP_NAME)/deploy/helm > /dev/null || exit 1
	@# Use helm upgrade --install to handle both fresh installs and updates gracefully
	@helm upgrade --install $(HELM_RELEASE_NAME) apps/$(APP_NAME)/deploy/helm \
		-f apps/$(APP_NAME)/deploy/helm/$(HELM_VALUES) \
		--namespace $(APP_NAME) --create-namespace
	@echo ""
	@echo "Deployment complete! Access the application:"
	@echo "  - HTTP: http://$(APP_NAME).k3d.local"
	@echo "  - Metrics: http://$(APP_NAME).k3d.local/metrics"
	@echo ""
	@echo "View deployment status:"
	@echo "  kubectl get deployments -n $(APP_NAME)"
	@echo "  kubectl get pods -n $(APP_NAME)"
	@echo "  kubectl get svc -n $(APP_NAME)"

deploy-all: $(APPS:%=deploy-%)

$(APPS:%=deploy-%):
	$(MAKE) deploy APP=$(@:deploy-%=%)

destroy-app:
	@echo "Uninstalling $(APP_NAME) from k3d cluster..."
	@helm uninstall $(HELM_RELEASE_NAME) -n $(APP_NAME) || true
	@kubectl delete namespace $(APP_NAME) --ignore-not-found
	@echo "Uninstall complete"

destroy-all-apps: $(APPS:%=destroy-%)

$(APPS:%=destroy-%):
	$(MAKE) destroy-app APP=$(@:destroy-%=%)

helm-lint:
	@echo "Linting Helm chart..."
	@helm lint apps/$(APP_NAME)/deploy/helm -f apps/$(APP_NAME)/deploy/helm/$(HELM_VALUES)
	@echo "Lint complete"

helm-validate:
	@echo "Validating Helm chart (dry-run)..."
	@echo "Using values file: apps/$(APP_NAME)/deploy/helm/$(HELM_VALUES)"
	@helm template $(HELM_RELEASE_NAME) apps/$(APP_NAME)/deploy/helm \
		-f apps/$(APP_NAME)/deploy/helm/$(HELM_VALUES) \
		--namespace $(APP_NAME)