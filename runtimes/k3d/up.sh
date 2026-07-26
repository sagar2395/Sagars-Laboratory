#!/usr/bin/env bash
set -euo pipefail

# Create a k3d cluster
# Expose HTTP/HTTPS ports on the host so that ingress rules using
# hostnames resolve to localhost and traffic reaches the load-balancer pod.
# Values come from .env (via labctl config) or environment variables.

CLUSTER_NAME="${1:-${CLUSTER_NAME:-flightdeck}}"
HTTP_PORT="${HTTP_PORT:-80}"
HTTPS_PORT="${HTTPS_PORT:-443}"
# Number of agent (worker) nodes. Multi-node by default so day-2 drills
# (node drain, rolling upgrade — task 060) have somewhere to reschedule pods.
AGENTS="${AGENTS:-2}"
# Optional k3s version pin (e.g. K3S_VERSION=v1.28.8-k3s1). Empty = k3d default.
# The cluster-upgrade-drill creates a cluster pinned to an older version, then
# rolls the agents to a newer one.
K3S_VERSION="${K3S_VERSION:-}"

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
create_args=(
  "$CLUSTER_NAME"
  --agents "$AGENTS"
  -p "${HTTP_PORT}:80@loadbalancer"
  -p "${HTTPS_PORT}:443@loadbalancer"
  --k3s-arg "--disable=traefik@server:*"
)
if [ -n "$K3S_VERSION" ]; then
  echo "Pinning k3s version to ${K3S_VERSION}"
  create_args+=(--image "rancher/k3s:${K3S_VERSION}")
fi

k3d cluster create "${create_args[@]}"

kubectl config use-context "k3d-$CLUSTER_NAME"
