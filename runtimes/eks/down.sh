#!/usr/bin/env bash
set -euo pipefail

# Destroy an EKS cluster using Terraform
# Usage: down.sh [cluster-name]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

CLUSTER_NAME="${1:-sagars-cluster}"
AWS_REGION="${AWS_REGION:-us-east-1}"
TF_DIR="${PROJECT_ROOT}/foundation/terraform/environments/dev"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Destroying EKS cluster '$CLUSTER_NAME'...${NC}"

# Check if Terraform state exists
if [ -f "$TF_DIR/terraform.tfstate" ]; then
    echo "[1/2] Running Terraform destroy..."
    terraform -chdir="$TF_DIR" destroy -auto-approve -input=false \
        -var="cluster_name=$CLUSTER_NAME" \
        -var="aws_region=$AWS_REGION" \
        -var="runtime=eks" || true
else
    echo "[1/2] No Terraform state found. Attempting direct EKS deletion..."
    eksctl delete cluster --name "$CLUSTER_NAME" --region "$AWS_REGION" --wait 2>/dev/null || \
        aws eks delete-cluster --name "$CLUSTER_NAME" --region "$AWS_REGION" 2>/dev/null || true
fi

# Clean up kubeconfig context
echo "[2/2] Cleaning up kubeconfig..."
kubectl config delete-context "arn:aws:eks:${AWS_REGION}:*:cluster/${CLUSTER_NAME}" 2>/dev/null || true
kubectl config unset "users.arn:aws:eks:${AWS_REGION}:*:cluster/${CLUSTER_NAME}" 2>/dev/null || true

echo -e "${GREEN}EKS cluster '$CLUSTER_NAME' has been destroyed.${NC}"
