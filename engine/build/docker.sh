#!/bin/bash

# Docker Build and Import Script
# Usage: docker.sh <app-name> [--import] [--cluster-name] [--profile]

set -e

APP_NAME="${1:?Error: APP_NAME not provided}"

# source any app-specific configuration (optional)
if [ -f "apps/${APP_NAME}/app.env" ]; then
    # shellcheck disable=SC1090
    set -a; . "apps/${APP_NAME}/app.env"; set +a
fi

CLUSTER_NAME="${CLUSTER_NAME:-sagars-cluster}"
PROFILE="${PROFILE:-k3d}"

echo "Building Docker image for ${APP_NAME}..."
docker build -t "${APP_NAME}:latest" "apps/${APP_NAME}/"

if [ "$2" == "--import" ] && [ "${PROFILE}" == "k3d" ]; then
    echo "Importing Docker image into k3d cluster '${CLUSTER_NAME}'..."
    k3d image import "${APP_NAME}:latest" -c "${CLUSTER_NAME}"
    echo "✓ Image imported successfully"
fi

echo "✓ Docker build complete"