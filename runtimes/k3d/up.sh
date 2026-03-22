#!/usr/bin/env bash
set -euo pipefail

# Create a k3d cluster
# Expose HTTP/HTTPS ports on the host so that ingress rules using
# hostnames resolve to localhost and traffic reaches the load-balancer pod.
# Values come from .env (via labctl config) or environment variables.

CLUSTER_NAME="${1:-${CLUSTER_NAME:-sagars-cluster}}"
HTTP_PORT="${HTTP_PORT:-80}"
HTTPS_PORT="${HTTPS_PORT:-443}"

# Skip if cluster already exists
if k3d cluster list "$CLUSTER_NAME" &>/dev/null; then
  echo "Cluster '$CLUSTER_NAME' already exists, skipping creation."
  kubectl config use-context "k3d-$CLUSTER_NAME"
  exit 0
fi

# Disable the bundled Traefik so we manage our own install in the traefik namespace.
# This prevents two competing Traefik instances from causing 404 errors.
k3d cluster create "$CLUSTER_NAME" --agents 2 \
  -p "${HTTP_PORT}:80@loadbalancer" \
  -p "${HTTPS_PORT}:443@loadbalancer" \
  --k3s-arg "--disable=traefik@server:*"

kubectl config use-context "k3d-$CLUSTER_NAME"
