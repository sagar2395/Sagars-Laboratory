#!/usr/bin/env bash
set -euo pipefail

# ECR Build Strategy — build and push to AWS Elastic Container Registry
# Usage: ecr.sh <app-name>
# Requires: aws cli configured, AWS_ECR_ACCOUNT_ID and AWS_REGION set

APP_NAME="${1:?Error: APP_NAME not provided}"

# Source app-specific configuration
if [ -f "apps/${APP_NAME}/app.env" ]; then
    set -a; . "apps/${APP_NAME}/app.env"; set +a
fi

AWS_REGION="${AWS_REGION:?AWS_REGION must be set (e.g. in runtimes/eks/runtime.env)}"
AWS_ECR_ACCOUNT_ID="${AWS_ECR_ACCOUNT_ID:?AWS_ECR_ACCOUNT_ID must be set}"
ECR_REPO_PREFIX="${AWS_ECR_REPO_PREFIX:-sagars-lab}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

ECR_REGISTRY="${AWS_ECR_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
FULL_IMAGE="${ECR_REGISTRY}/${ECR_REPO_PREFIX}/${APP_NAME}:${IMAGE_TAG}"

echo "Building and pushing to ECR..."
echo "  Registry: ${ECR_REGISTRY}"
echo "  Image:    ${FULL_IMAGE}"
echo ""

# Login to ECR
echo "[1/3] Logging in to ECR..."
aws ecr get-login-password --region "$AWS_REGION" | \
    docker login --username AWS --password-stdin "$ECR_REGISTRY"

# Build image
echo "[2/3] Building Docker image..."
docker build -t "${FULL_IMAGE}" "apps/${APP_NAME}/"

# Also tag as latest if not already
if [ "$IMAGE_TAG" != "latest" ]; then
    docker tag "${FULL_IMAGE}" "${ECR_REGISTRY}/${ECR_REPO_PREFIX}/${APP_NAME}:latest"
fi

# Push to ECR
echo "[3/3] Pushing to ECR..."
docker push "${FULL_IMAGE}"
if [ "$IMAGE_TAG" != "latest" ]; then
    docker push "${ECR_REGISTRY}/${ECR_REPO_PREFIX}/${APP_NAME}:latest"
fi

echo ""
echo "✓ Image pushed to ECR: ${FULL_IMAGE}"
echo ""
echo "To deploy, update your Helm values:"
echo "  image.repository: ${ECR_REGISTRY}/${ECR_REPO_PREFIX}/${APP_NAME}"
echo "  image.tag: ${IMAGE_TAG}"
