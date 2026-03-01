#!/bin/bash

# Wrapper for deploy strategies.  Loads the app's configuration and dispatches
# to the appropriate strategy script in engine/deploy/<strategy>.sh.
# Usage: deploy.sh <command> <app-name> [options passed to strategy script]

set -euo pipefail

COMMAND="${1:?Usage: $0 <command> <app-name>}"
APP_NAME="${2:?Usage: $0 <command> <app-name>}"
shift 2 || true  # Remove COMMAND and APP_NAME, pass remaining args to strategy

# Load app configuration to get DEPLOY_STRATEGY
if [ -f "apps/${APP_NAME}/app.env" ]; then
    set -a; . "apps/${APP_NAME}/app.env"; set +a
fi

# Use explicit override if provided, otherwise use app.env default, otherwise fail
DEPLOY_STRATEGY="${DEPLOY_STRATEGY:?app.env must define DEPLOY_STRATEGY or pass DEPLOY_STRATEGY=... on command line}"

echo "[${COMMAND}] app=${APP_NAME} strategy=${DEPLOY_STRATEGY}"
bash "engine/deploy/${DEPLOY_STRATEGY}.sh" "${COMMAND}" "${APP_NAME}" "$@"
