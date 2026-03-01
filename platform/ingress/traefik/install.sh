#!/bin/bash
set -e

NAMESPACE=traefik

echo "Installing Traefik..."

helm repo add traefik https://traefik.github.io/charts >/dev/null 2>&1 || true
helm repo update

helm upgrade --install traefik traefik/traefik \
  --namespace $NAMESPACE \
  --create-namespace \
  --set service.type=LoadBalancer \
  --set dashboard.enabled=true

echo "Waiting for Traefik to be ready..."
kubectl rollout status deployment/traefik -n $NAMESPACE --timeout=120s

echo "Traefik installed successfully."