# Application Build and Deployment Targets
# ### each app provides its own configuration (apps/<name>/app.env) containing
# BUILD_STRATEGY, DEPLOY_STRATEGY and any strategy-specific variables.  make
# targets simply load that file and invoke the corresponding script.

APPS := $(shell ls apps)

# helper macro to load application-specific variables; exported in a subshell.
define load_app_config
	@if [ -n "$(APP_NAME)" ] && [ -f "apps/$(APP_NAME)/app.env" ]; then \
		echo "Loading configuration for $(APP_NAME)"; \
		set -a; . "apps/$(APP_NAME)/app.env"; set +a; \
	fi
endef

.PHONY: build local-run deploy destroy-app lint validate \
        deploy-all destroy-all-apps

# build target dispatches to the chosen build strategy
build:
	@$(call load_app_config)
	@echo "[build] app=$(APP_NAME) strategy=$(BUILD_STRATEGY)"
	@bash engine/build/$(BUILD_STRATEGY).sh $(APP_NAME)

# run locally if the app is a binary
local-run:
	@echo "Running $(APP_NAME) locally..."
	@cd apps/$(APP_NAME) && go run main.go

# deploy/destroy are completely strategy-driven
deploy:
	@$(call load_app_config)
	@echo "[deploy] app=$(APP_NAME) strategy=$(DEPLOY_STRATEGY)"
	@bash engine/deploy/$(DEPLOY_STRATEGY).sh deploy $(APP_NAME)

destroy-app:
	@$(call load_app_config)
	@echo "[destroy] app=$(APP_NAME) strategy=$(DEPLOY_STRATEGY)"
	@bash engine/deploy/$(DEPLOY_STRATEGY).sh destroy $(APP_NAME)

lint:
	@$(call load_app_config)
	@bash engine/deploy/$(DEPLOY_STRATEGY).sh lint $(APP_NAME)

validate:
	@$(call load_app_config)
	@bash engine/deploy/$(DEPLOY_STRATEGY).sh validate $(APP_NAME)

# Bulk operations for all apps
deploy-all: $(APPS:%=deploy-%)

$(APPS:%=deploy-%):
	@$(MAKE) deploy APP_NAME=$(@:deploy-%=%)

destroy-all-apps: $(APPS:%=destroy-%)

$(APPS:%=destroy-%):
	@$(MAKE) destroy-app APP_NAME=$(@:destroy-%=%)