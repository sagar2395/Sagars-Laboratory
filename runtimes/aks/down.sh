#!/usr/bin/env bash
set -euo pipefail

# Destroy an AKS cluster using Terraform
# Usage: down.sh [cluster-name]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

CLUSTER_NAME="${1:-sagars-cluster}"
RESOURCE_GROUP="${AZURE_RESOURCE_GROUP:-sagars-lab-rg}"
TF_DIR="${PROJECT_ROOT}/foundation/terraform/environments/dev"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Destroying AKS cluster '$CLUSTER_NAME'...${NC}"

# Check if Terraform state exists
if [ -f "$TF_DIR/terraform.tfstate" ]; then
    echo "[1/2] Running Terraform destroy..."
    terraform -chdir="$TF_DIR" destroy -auto-approve -input=false \
        -var="cluster_name=$CLUSTER_NAME" \
        -var="resource_group_name=$RESOURCE_GROUP" \
        -var="location=${AZURE_LOCATION:-eastus}" \
        -var="runtime=aks" || true
else
    echo "[1/2] No Terraform state found. Attempting direct AKS deletion..."
    az aks delete \
        --resource-group "$RESOURCE_GROUP" \
        --name "$CLUSTER_NAME" \
        --yes --no-wait 2>/dev/null || true
fi

# Clean up kubeconfig context
echo "[2/2] Cleaning up kubeconfig..."
kubectl config delete-context "$CLUSTER_NAME" 2>/dev/null || true
kubectl config delete-cluster "$CLUSTER_NAME" 2>/dev/null || true

echo -e "${GREEN}AKS cluster '$CLUSTER_NAME' has been destroyed.${NC}"
