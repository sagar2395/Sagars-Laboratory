# Cloud Runtimes

This guide covers deploying the lab to cloud Kubernetes clusters on Azure (AKS) and AWS (EKS) using Terraform.

## Overview

The lab supports these runtime profiles:

| Profile | Cluster | Prerequisites | Cost |
|---------|---------|--------------|------|
| `k3d` | Local k3d cluster | Docker only | Free |
| `kind` | Local kind cluster (headless, CI-friendly) | Docker + kind | Free |
| `aks` | Azure Kubernetes Service | Azure subscription + CLI | Pay-per-use |
| `eks` | AWS Elastic Kubernetes Service | AWS account + CLI | Pay-per-use |
| `gke` | Google Kubernetes Engine | GCP project + gcloud | Pay-per-use |
| `incluster` | The hosting cluster itself (team-mode server) | — | — |

`kind` mirrors k3d's host-port exposure (control-plane `extraPortMappings` for
80/443 + `ingress-ready` label), so platform scripts don't care which local
runtime is underneath. It is the runtime used by the nightly e2e CI job
(`.github/workflows/e2e-kind.yaml`).

Each runtime defines its own environment in `runtimes/<profile>/runtime.env`:

```bash
# Example: runtimes/aks/runtime.env
INGRESS_CLASS=nginx
STORAGE_CLASS=managed-csi
DOMAIN_SUFFIX=sagarslab.io
REGISTRY_TYPE=acr
```

## Prerequisites

### Local (kind)

```bash
# Install docker + kind + kubectl (kind: https://kind.sigs.k8s.io)
export PROFILE=kind
labctl runtime up            # creates a control-plane + 1 worker kind cluster
# Optional knobs (env): AGENTS=2 (workers), KIND_NODE_IMAGE=kindest/node:v1.29.4,
#                       HTTP_PORT / HTTPS_PORT (host ports for ingress)
labctl runtime down
```

### Google (GKE)

```bash
# Install tools, then authenticate
make setup-tools PROFILE=gke   # or install gcloud + terraform + kubectl
gcloud auth login
gcloud auth application-default login

# Set the project in runtimes/gke/runtime.env (or the env):
#   GCP_PROJECT=my-project-id
#   GCP_REGION=us-central1
```

### Azure (AKS)

```bash
# Install tools
make setup-tools PROFILE=aks

# Login
az login
az account set --subscription <subscription-id>

# Create resource group (if needed)
az group create --name sagars-lab-rg --location eastus
```

### AWS (EKS)

```bash
# Install tools
make setup-tools PROFILE=eks

# Configure credentials
aws configure
# Or use environment variables:
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1
```

## Configuration

Edit `.env` to switch profiles:

```bash
PROFILE=aks    # or eks
CLUSTER_NAME=sagars-cluster
```

For Azure, add to `.env`:

```bash
AZURE_RESOURCE_GROUP=sagars-lab-rg
AZURE_LOCATION=eastus
AZURE_ACR_NAME=sagarslab
```

For AWS, add to `.env`:

```bash
AWS_REGION=us-east-1
AWS_ECR_ACCOUNT_ID=123456789012
AWS_ECR_REPO_PREFIX=sagars-lab
```

For GCP, add to `.env` (or `runtimes/gke/runtime.env`):

```bash
GCP_PROJECT=my-project-id
GCP_REGION=us-central1
GAR_REPO=sagars-lab
```

## Terraform Modules

Infrastructure is defined in `foundation/terraform/`:

```
foundation/terraform/
  modules/
    aks/           # AKS cluster + Log Analytics + ACR
      main.tf
      variables.tf
      outputs.tf
    eks/           # VPC + EKS cluster + node group + ECR
      main.tf
      variables.tf
      outputs.tf
    gke/           # Regional GKE cluster + node pool + Artifact Registry
      main.tf
      variables.tf
      outputs.tf
  environments/
    dev/           # Small cluster (2 nodes, basic VMs)
      main.tf
    staging/       # Larger cluster (3+ nodes, autoscaling)
      main.tf
```

### AKS Module Resources

- AKS cluster with Calico network policy and SystemAssigned identity
- Log Analytics workspace for container insights
- Azure Container Registry (optional)
- AcrPull role assignment for the cluster

### EKS Module Resources

- VPC with 2 public + 2 private subnets
- Internet Gateway + NAT Gateway
- IAM roles for cluster and node group
- EKS cluster + managed node group
- ECR repositories with lifecycle policies (keep last 10 images)

### GKE Module Resources

- Regional GKE cluster on the REGULAR release channel (default node pool removed)
- A separately-managed node pool with autoscaling (auto-repair + auto-upgrade)
- Artifact Registry (Docker) repository for app images (optional)

The root `environments/dev/main.tf` switches on `var.runtime` (`aks` | `eks` |
`gke`); only the matching module evaluates (the others have `count = 0`), so the
unused cloud provider is never authenticated.

> **Verify-once caveat (like AKS/EKS tasks 038/039):** the GKE module has not yet
> been applied against a real GCP project. Provision once, record the cost, then
> tear down — `gke` is unverified until then.

### Environment Sizes

| Setting | Dev | Staging |
|---------|-----|---------|
| AKS VM | Standard_B2s | Standard_B4ms |
| EKS Instance | t3.medium | t3.large |
| Nodes | 2 | 3 (autoscaling 3-8) |
| ACR/ECR | Basic/Standard | Basic/Standard |

## Provisioning

### Using Make

```bash
# Initialize Terraform
make terraform-init TF_ENV=dev

# Preview changes
make terraform-plan TF_ENV=dev

# Apply (creates cluster)
make terraform-apply TF_ENV=dev

# Check outputs
make terraform-output TF_ENV=dev
```

Pass the runtime variable:

```bash
make terraform-plan TF_ENV=dev TF_VARS='-var runtime=aks'
make terraform-apply TF_ENV=dev TF_VARS='-var runtime=aks'
```

### Using Runtime Scripts

The runtime scripts wrap Terraform:

```bash
# AKS
./runtimes/aks/up.sh     # az login check -> terraform apply -> get-credentials
./runtimes/aks/down.sh   # terraform destroy -> kubeconfig cleanup

# EKS
./runtimes/eks/up.sh     # aws sts check -> terraform apply -> update-kubeconfig
./runtimes/eks/down.sh   # terraform destroy -> kubeconfig cleanup

# GKE
./runtimes/gke/up.sh     # gcloud auth check -> terraform apply -> get-credentials
./runtimes/gke/down.sh   # terraform destroy

# kind (local, no Terraform)
./runtimes/kind/up.sh    # docker check -> kind create cluster (ports + ingress-ready)
./runtimes/kind/down.sh  # kind delete cluster
```

### Using labctl

```bash
export PROFILE=aks       # or kind | eks | gke
labctl runtime up        # calls runtimes/<profile>/up.sh
labctl runtime status
labctl runtime down
```

## State Management

By default, Terraform stores state in a local `terraform.tfstate` file. This is
fine for solo k3d work, but for cloud deployments (AKS/EKS) you need remote state
so that CI and multiple operators share a consistent view.

### How it works

Backend configuration lives in `backend.tf` (gitignored) next to each environment's
`main.tf`. The repo ships `backend.tf.example` in each environment showing both
provider options. Pick the one that matches your runtime and copy it.

### One-time setup: AzureRM backend (for AKS)

**1. Create the storage resources (run once):**

```bash
az group create -n sagars-lab-tfstate -l eastus
az storage account create -n sagarslabtfstate -g sagars-lab-tfstate -l eastus --sku Standard_LRS
az storage container create -n tfstate --account-name sagarslabtfstate
```

**2. Create `backend.tf` from the example:**

```bash
cd foundation/terraform/environments/dev
cp backend.tf.example backend.tf
# Uncomment the azurerm block, comment out the s3 block
```

The `backend.tf` for dev should look like:

```hcl
terraform {
  backend "azurerm" {
    resource_group_name  = "sagars-lab-tfstate"
    storage_account_name = "sagarslabtfstate"
    container_name       = "tfstate"
    key                  = "dev/terraform.tfstate"
  }
}
```

**3. Initialise (migrates any local state into the backend):**

```bash
terraform init -migrate-state
```

### One-time setup: S3 backend (for EKS)

**1. Create the S3 bucket and DynamoDB lock table:**

```bash
aws s3api create-bucket \
  --bucket sagars-lab-tfstate \
  --region us-east-1 \
  --create-bucket-configuration LocationConstraint=us-east-1

aws s3api put-bucket-versioning \
  --bucket sagars-lab-tfstate \
  --versioning-configuration Status=Enabled

aws dynamodb create-table \
  --table-name sagars-lab-tfstate-lock \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

**2. Create `backend.tf` from the example:**

```bash
cd foundation/terraform/environments/dev
cp backend.tf.example backend.tf
# Uncomment the s3 block, comment out the azurerm block
```

**3. Initialise:**

```bash
terraform init -migrate-state
```

### Local-only use (no backend.tf)

Without a `backend.tf`, Terraform falls back to local state. The `*.tfstate` files
are gitignored, so each operator has independent state. This is acceptable for
one-off personal experiments on k3d, **not** for any shared or CI workflow.

## Building and Deploying to Cloud

### Image Registry

Cloud runtimes use their own container registries instead of k3d's local import:

```bash
# Set build strategy in app.env
BUILD_STRATEGY=acr    # for Azure
BUILD_STRATEGY=ecr    # for AWS
```

Build scripts handle authentication:

```bash
# ACR: engine/build/acr.sh
az acr login --name $AZURE_ACR_NAME
docker build + docker push

# ECR: engine/build/ecr.sh
aws ecr get-login-password | docker login
docker build + docker push
```

### Helm Values

Use the cloud values profile:

```bash
# In app.env
HELM_VALUES=values-cloud.yaml
```

The `values-cloud.yaml` files configure:
- `className: nginx` (cloud ingress controller)
- `pullPolicy: Always` (pull from registry)
- Appropriate resource requests/limits
- Liveness and readiness probes
- Pod anti-affinity for spread across nodes

Update the image repository in `values-cloud.yaml` to match your registry:

```yaml
# AKS
image:
  repository: sagarslab.azurecr.io/go-api

# EKS
image:
  repository: 123456789012.dkr.ecr.us-east-1.amazonaws.com/sagars-lab/go-api
```

### Platform Components

Cloud runtimes use Nginx instead of Traefik for ingress:

```bash
# In .env
INGRESS_PROVIDER=nginx
```

All scenarios work identically on cloud runtimes. The scenario engine resolves `{{.DomainSuffix}}` from the active runtime's `runtime.env`.

## Cost Considerations

| Resource | Approximate Cost |
|----------|-----------------|
| AKS cluster (control plane) | Free |
| AKS Standard_B2s nodes (x2) | ~$60/month |
| Azure ACR Basic | ~$5/month |
| EKS cluster (control plane) | ~$73/month |
| EKS t3.medium nodes (x2) | ~$60/month |
| EKS NAT Gateway | ~$32/month |
| AWS ECR | ~$0 (pay per storage) |

**Tear down when not in use** to avoid charges:

```bash
make terraform-destroy TF_ENV=dev TF_VARS='-var runtime=aks'
```

## Troubleshooting

### AKS: Cannot pull images

```bash
# Verify ACR is attached
az aks check-acr --name sagars-cluster --resource-group sagars-lab-rg --acr sagarslab
```

### EKS: Nodes not joining

```bash
# Check node group status
aws eks describe-nodegroup --cluster-name sagars-cluster --nodegroup-name default
# Check IAM role trust
aws iam get-role --role-name sagars-cluster-node-role
```

### General: Wrong kubeconfig context

```bash
# AKS
az aks get-credentials --resource-group sagars-lab-rg --name sagars-cluster --overwrite
# EKS
aws eks update-kubeconfig --name sagars-cluster --region us-east-1
# Verify
kubectl config current-context
```
