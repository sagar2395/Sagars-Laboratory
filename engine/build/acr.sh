#!/usr/bin/env bash
set -euo pipefail

# ACR Build Strategy — build and push to Azure Container Registry
# Usage: acr.sh <app-name>
# Requires: az cli authenticated, ACR_NAME set in env

APP_NAME="${1:?Error: APP_NAME not provided}"

# Source app-specific configuration
if [ -f "apps/${APP_NAME}/app.env" ]; then
    set -a; . "apps/${APP_NAME}/app.env"; set +a
fi

ACR_NAME="${AZURE_ACR_NAME:?ACR_NAME must be set (e.g. in runtimes/aks/runtime.env)}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

ACR_LOGIN_SERVER="${ACR_NAME}.azurecr.io"
FULL_IMAGE="${ACR_LOGIN_SERVER}/${APP_NAME}:${IMAGE_TAG}"

echo "Building and pushing to ACR..."
echo "  Registry: ${ACR_LOGIN_SERVER}"
echo "  Image:    ${FULL_IMAGE}"
echo ""

# Login to ACR
echo "[1/3] Logging in to ACR..."
az acr login --name "$ACR_NAME"

# Build image
echo "[2/3] Building Docker image..."
docker build -t "${FULL_IMAGE}" "apps/${APP_NAME}/"

# Also tag as latest if not already
if [ "$IMAGE_TAG" != "latest" ]; then
    docker tag "${FULL_IMAGE}" "${ACR_LOGIN_SERVER}/${APP_NAME}:latest"
fi

# Push to ACR
echo "[3/3] Pushing to ACR..."
docker push "${FULL_IMAGE}"
if [ "$IMAGE_TAG" != "latest" ]; then
    docker push "${ACR_LOGIN_SERVER}/${APP_NAME}:latest"
fi

echo ""
echo "✓ Image pushed to ACR: ${FULL_IMAGE}"
echo ""
echo "To deploy, update your Helm values:"
echo "  image.repository: ${ACR_LOGIN_SERVER}/${APP_NAME}"
echo "  image.tag: ${IMAGE_TAG}"
