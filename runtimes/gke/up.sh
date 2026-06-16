#!/usr/bin/env bash
set -euo pipefail

# Create a GKE cluster using Terraform, then fetch kubeconfig.
# Usage: up.sh <cluster-name>
# Prerequisites: gcloud authenticated (gcloud auth login + application-default),
# GCP_PROJECT set, Terraform available.
#
# Mirrors runtimes/aks/up.sh and runtimes/eks/up.sh; same caveat as 038/039 —
# verify against a real project once, record cost, then tear down.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

CLUSTER_NAME="${1:-${CLUSTER_NAME:-sagars-cluster}}"
GCP_PROJECT="${GCP_PROJECT:-}"
GCP_REGION="${GCP_REGION:-us-central1}"
TF_DIR="${PROJECT_ROOT}/foundation/terraform/environments/dev"

echo "=== GKE Cluster Setup ==="
echo "  Cluster:  $CLUSTER_NAME"
echo "  Project:  ${GCP_PROJECT:-<unset>}"
echo "  Region:   $GCP_REGION"
echo ""

if [ -z "$GCP_PROJECT" ]; then
    echo "ERROR: GCP_PROJECT is not set. Set it in runtimes/gke/runtime.env or the env." >&2
    exit 1
fi

for bin in gcloud terraform kubectl; do
    if ! command -v "$bin" >/dev/null 2>&1; then
        echo "ERROR: '$bin' is required but not installed." >&2
        exit 1
    fi
done

# Verify gcloud is authenticated.
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" 2>/dev/null | grep -q .; then
    echo "ERROR: Not authenticated to GCP. Run 'gcloud auth login' and" >&2
    echo "       'gcloud auth application-default login' first." >&2
    exit 1
fi

echo "[1/3] Running Terraform init..."
terraform -chdir="$TF_DIR" init -input=false

echo "[2/3] Running Terraform apply (GKE)..."
terraform -chdir="$TF_DIR" apply -auto-approve -input=false \
    -var="runtime=gke" \
    -var="cluster_name=$CLUSTER_NAME" \
    -var="gcp_project=$GCP_PROJECT" \
    -var="gcp_region=$GCP_REGION"

echo "[3/3] Fetching kubeconfig..."
gcloud container clusters get-credentials "$CLUSTER_NAME" \
    --region "$GCP_REGION" \
    --project "$GCP_PROJECT"

echo ""
echo "=== GKE cluster '$CLUSTER_NAME' is ready ==="
echo "  kubectl context: $(kubectl config current-context 2>/dev/null || echo '(unknown)')"
kubectl get nodes -o wide 2>/dev/null || echo "  (waiting for nodes...)"
