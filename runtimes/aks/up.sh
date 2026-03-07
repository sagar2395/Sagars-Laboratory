#!/usr/bin/env bash
set -euo pipefail

# Create an AKS cluster using Terraform
# Usage: up.sh <cluster-name>
# Prerequisites: az cli logged in, Terraform initialized

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

CLUSTER_NAME="${1:-sagars-cluster}"
RESOURCE_GROUP="${AZURE_RESOURCE_GROUP:-sagars-lab-rg}"
LOCATION="${AZURE_LOCATION:-eastus}"
TF_DIR="${PROJECT_ROOT}/foundation/terraform/environments/dev"

echo "=== AKS Cluster Setup ==="
echo "  Cluster:        $CLUSTER_NAME"
echo "  Resource Group:  $RESOURCE_GROUP"
echo "  Location:        $LOCATION"
echo ""

# Verify az CLI is authenticated
if ! az account show >/dev/null 2>&1; then
    echo "ERROR: Not logged in to Azure. Run 'az login' first."
    exit 1
fi

# Create resource group if it doesn't exist
echo "[1/4] Ensuring resource group '$RESOURCE_GROUP' exists..."
az group create --name "$RESOURCE_GROUP" --location "$LOCATION" --output none 2>/dev/null || true

# Run Terraform
echo "[2/4] Running Terraform init..."
terraform -chdir="$TF_DIR" init -input=false

echo "[3/4] Running Terraform apply (AKS)..."
terraform -chdir="$TF_DIR" apply -auto-approve -input=false \
    -var="cluster_name=$CLUSTER_NAME" \
    -var="resource_group_name=$RESOURCE_GROUP" \
    -var="location=$LOCATION" \
    -var="runtime=aks"

# Get kubeconfig
echo "[4/4] Fetching kubeconfig..."
az aks get-credentials \
    --resource-group "$RESOURCE_GROUP" \
    --name "$CLUSTER_NAME" \
    --overwrite-existing

echo ""
echo "=== AKS cluster '$CLUSTER_NAME' is ready ==="
echo "  kubectl context: $(kubectl config current-context)"
echo "  Nodes:"
kubectl get nodes -o wide 2>/dev/null || echo "  (waiting for nodes...)"
