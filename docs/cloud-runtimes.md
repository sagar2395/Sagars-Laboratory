# Cloud Runtimes

This guide covers deploying the lab to cloud Kubernetes clusters on Azure (AKS) and AWS (EKS) using Terraform.

## Overview

The lab supports three runtime profiles:

| Profile | Cluster | Prerequisites | Cost |
|---------|---------|--------------|------|
| `k3d` | Local k3d cluster | Docker only | Free |
| `aks` | Azure Kubernetes Service | Azure subscription + CLI | Pay-per-use |
| `eks` | AWS Elastic Kubernetes Service | AWS account + CLI | Pay-per-use |

Each runtime defines its own environment in `runtimes/<profile>/runtime.env`:

```bash
# Example: runtimes/aks/runtime.env
INGRESS_CLASS=nginx
STORAGE_CLASS=managed-csi
DOMAIN_SUFFIX=sagarslab.io
REGISTRY_TYPE=acr
```

## Prerequisites

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
```

### Using labctl

```bash
export PROFILE=aks
labctl runtime up      # calls runtimes/aks/up.sh
labctl runtime status
labctl runtime down
```

## Remote State

For team use, enable remote Terraform state. Both environments have commented-out backend blocks:

### Azure Blob Storage

```hcl
# In foundation/terraform/environments/dev/main.tf
backend "azurerm" {
  resource_group_name  = "terraform-state-rg"
  storage_account_name = "sagarslabstate"
  container_name       = "tfstate"
  key                  = "dev.terraform.tfstate"
}
```

Create the storage account first:

```bash
az group create -n terraform-state-rg -l eastus
az storage account create -n sagarslabstate -g terraform-state-rg -l eastus --sku Standard_LRS
az storage container create -n tfstate --account-name sagarslabstate
```

### AWS S3

```hcl
# In foundation/terraform/environments/dev/main.tf
backend "s3" {
  bucket = "sagars-lab-tfstate"
  key    = "dev/terraform.tfstate"
  region = "us-east-1"
}
```

Create the bucket first:

```bash
aws s3 mb s3://sagars-lab-tfstate --region us-east-1
```

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
