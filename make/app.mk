# Application Build and Deployment Targets
# ### each app provides its own configuration (apps/<name>/app.env) containing
# BUILD_STRATEGY, DEPLOY_STRATEGY and any strategy-specific variables.  make
# targets simply load that file and invoke the corresponding script.

APPS := $(shell ls apps)


.PHONY: build local-run deploy destroy-app lint validate \
        deploy-all destroy-all-apps

# build target dispatches to the chosen build strategy
build:
	@bash engine/build.sh $(APP_NAME)

# run locally if the app is a binary
local-run:
	@bash engine/run.sh $(APP_NAME)

# deploy/destroy are completely strategy-driven
deploy:
	@bash engine/deploy.sh deploy $(APP_NAME)

destroy-app:
	@bash engine/deploy.sh destroy $(APP_NAME)

lint:
	@bash engine/deploy.sh lint $(APP_NAME)

validate:
	@bash engine/deploy.sh validate $(APP_NAME)

# Bulk operations for all apps
deploy-all: $(APPS:%=deploy-%)

$(APPS:%=deploy-%):
	@$(MAKE) deploy APP_NAME=$(@:deploy-%=%)

destroy-all-apps: $(APPS:%=destroy-%)

$(APPS:%=destroy-%):
	@$(MAKE) destroy-app APP_NAME=$(@:destroy-%=%)