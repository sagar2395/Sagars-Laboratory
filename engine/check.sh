#!/bin/bash

# Encapsulate all kubectl/helm/docker checks previously scattered in makefiles.
# Invoke as:
#   engine/check.sh tools
#   engine/check.sh cluster
#   engine/check.sh ingress

set -euo pipefail

cmd=${1:-}
case "$cmd" in
  tools)
    echo "Checking required command-line tools..."
    for bin in kubectl helm docker k3d; do
      if ! command -v "$bin" >/dev/null 2>&1; then
        echo "ERROR: $bin not found in PATH" >&2
        exit 1
      fi
    done
    echo "All required tools are available."
    ;;
  cluster)
    echo "Verifying access to Kubernetes cluster..."
    kubectl version --client --short >/dev/null 2>&1 || { echo "Cannot query cluster" >&2; exit 1; }
    echo "Current context: $(kubectl config current-context)"
    ;;
  ingress)
    echo "Checking platform ingress controller..."
    if ! kubectl get pods --all-namespaces -l app=traefik 2>/dev/null | grep -q Running; then
      echo "Traefik pods are not running" >&2
      exit 1
    fi
    echo "Traefik appears to be running."
    ;;
  *)
    echo "Usage: $0 {tools|cluster|ingress}" >&2
    exit 1
    ;;
esac
