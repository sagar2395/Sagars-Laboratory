#!/usr/bin/env bash
set -euo pipefail

# Create an EKS cluster using Terraform
# Usage: up.sh <cluster-name>
# Prerequisites: aws cli configured, Terraform initialized

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

CLUSTER_NAME="${1:-sagars-cluster}"
AWS_REGION="${AWS_REGION:-us-east-1}"
TF_DIR="${PROJECT_ROOT}/foundation/terraform/environments/dev"

echo "=== EKS Cluster Setup ==="
echo "  Cluster:  $CLUSTER_NAME"
echo "  Region:   $AWS_REGION"
echo ""

# Verify AWS CLI is configured
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "ERROR: AWS CLI is not configured. Run 'aws configure' first."
    exit 1
fi

# Run Terraform
echo "[1/3] Running Terraform init..."
terraform -chdir="$TF_DIR" init -input=false

echo "[2/3] Running Terraform apply (EKS)..."
terraform -chdir="$TF_DIR" apply -auto-approve -input=false \
    -var="cluster_name=$CLUSTER_NAME" \
    -var="aws_region=$AWS_REGION" \
    -var="runtime=eks"

# Get kubeconfig
echo "[3/3] Updating kubeconfig..."
aws eks update-kubeconfig \
    --region "$AWS_REGION" \
    --name "$CLUSTER_NAME"

echo ""
echo "=== EKS cluster '$CLUSTER_NAME' is ready ==="
echo "  kubectl context: $(kubectl config current-context)"
echo "  Nodes:"
kubectl get nodes -o wide 2>/dev/null || echo "  (waiting for nodes...)"
