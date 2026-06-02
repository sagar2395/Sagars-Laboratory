#!/usr/bin/env bash
set -euo pipefail

# Create a k3d cluster
# Expose HTTP/HTTPS ports on the host so that ingress rules using
# hostnames resolve to localhost and traffic reaches the load-balancer pod.
# Values come from .env (via labctl config) or environment variables.

CLUSTER_NAME="${1:-${CLUSTER_NAME:-sagars-cluster}}"
HTTP_PORT="${HTTP_PORT:-80}"
HTTPS_PORT="${HTTPS_PORT:-443}"

# ---------------------------------------------------------------------------
# Docker daemon readiness — auto-start Colima on macOS if needed
# ---------------------------------------------------------------------------
ensure_docker() {
    if docker info &>/dev/null; then
        return 0
    fi

    echo "Docker daemon not reachable. Checking for a container runtime..."

    if command -v colima &>/dev/null; then
        echo "Starting Colima..."
        colima start
        local retries=0
        until docker info &>/dev/null; do
            retries=$((retries + 1))
            if [ "$retries" -ge 30 ]; then
                echo "ERROR: Docker daemon still not reachable after 30s." >&2
                echo "       Run 'colima status' for details." >&2
                exit 1
            fi
            sleep 1
        done
        echo "Colima started and Docker daemon is ready."
        return 0
    fi

    echo "ERROR: Docker daemon is not running and no supported runtime was found." >&2
    echo "  macOS options (pick one):" >&2
    echo "    colima:         brew install colima && colima start" >&2
    echo "    Docker Desktop: https://www.docker.com/products/docker-desktop" >&2
    echo "    OrbStack:       https://orbstack.dev" >&2
    exit 1
}

ensure_docker

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
