#!/bin/bash
set -e

NAMESPACE=traefik

echo "Installing Traefik..."

helm repo add traefik https://traefik.github.io/charts >/dev/null 2>&1 || true
helm repo update

# use a values file for configurability; default values live in this repo
helm upgrade --install traefik traefik/traefik \
  --namespace $NAMESPACE \
  --create-namespace \
  -f "$(dirname "$0")/values.yaml" \
  --set service.type=LoadBalancer \
  --set api.dashboard=true  # enable the web dashboard via the API settings

echo "Waiting for Traefik to be ready..."
kubectl rollout status deployment/traefik -n $NAMESPACE --timeout=120s

echo "Traefik installed successfully."