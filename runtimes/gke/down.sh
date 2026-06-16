#!/usr/bin/env bash
set -euo pipefail

# Destroy the GKE cluster via Terraform. Mirrors aks/eks down.sh.
# Usage: down.sh <cluster-name>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

CLUSTER_NAME="${1:-${CLUSTER_NAME:-sagars-cluster}}"
GCP_PROJECT="${GCP_PROJECT:-}"
GCP_REGION="${GCP_REGION:-us-central1}"
TF_DIR="${PROJECT_ROOT}/foundation/terraform/environments/dev"

if [ -z "$GCP_PROJECT" ]; then
    echo "ERROR: GCP_PROJECT is not set." >&2
    exit 1
fi

if ! command -v terraform >/dev/null 2>&1; then
    echo "ERROR: 'terraform' is required but not installed." >&2
    exit 1
fi

echo "Destroying GKE cluster '$CLUSTER_NAME' (project ${GCP_PROJECT}, region ${GCP_REGION})..."
terraform -chdir="$TF_DIR" destroy -auto-approve -input=false \
    -var="runtime=gke" \
    -var="cluster_name=$CLUSTER_NAME" \
    -var="gcp_project=$GCP_PROJECT" \
    -var="gcp_region=$GCP_REGION"

echo "GKE cluster '$CLUSTER_NAME' destroyed."
