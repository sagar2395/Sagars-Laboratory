#!/bin/bash

# Wrapper for build strategies.  Loads the app's configuration and dispatches
# to the appropriate strategy script in engine/build/<strategy>.sh.
# Usage: build.sh <app-name> [options passed to strategy script]

set -euo pipefail

APP_NAME="${1:?Usage: $0 <app-name> [strategy-options]}"
shift || true  # Remove APP_NAME, pass remaining args to strategy script

# Load app configuration to get BUILD_STRATEGY
if [ -f "apps/${APP_NAME}/app.env" ]; then
    set -a; . "apps/${APP_NAME}/app.env"; set +a
fi

# Use explicit override if provided, otherwise use app.env default, otherwise fail
BUILD_STRATEGY="${BUILD_STRATEGY:?app.env must define BUILD_STRATEGY or pass BUILD_STRATEGY=... on command line}"

echo "[build] app=${APP_NAME} strategy=${BUILD_STRATEGY}"
bash "engine/build/${BUILD_STRATEGY}.sh" "${APP_NAME}" "$@"
