# Staging environment
# Uses larger nodes and autoscaling compared to dev.

terraform {
  required_version = ">= 1.5.0"

  # Uncomment to use remote state:
  # backend "azurerm" {
  #   resource_group_name  = "terraform-state-rg"
  #   storage_account_name = "sagarslabstate"
  #   container_name       = "tfstate"
  #   key                  = "staging.terraform.tfstate"
  # }
  # backend "s3" {
  #   bucket = "sagars-lab-tfstate"
  #   key    = "staging/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

variable "runtime" {
  description = "Which cloud runtime to provision: aks or eks"
  type        = string

  validation {
    condition     = contains(["aks", "eks"], var.runtime)
    error_message = "runtime must be 'aks' or 'eks'"
  }
}

variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "sagars-cluster-staging"
}

variable "resource_group_name" {
  description = "Azure resource group name (AKS only)"
  type        = string
  default     = "sagars-lab-staging-rg"
}

variable "location" {
  description = "Azure region (AKS only)"
  type        = string
  default     = "eastus"
}

variable "aws_region" {
  description = "AWS region (EKS only)"
  type        = string
  default     = "us-east-1"
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    project     = "sagars-laboratory"
    environment = "staging"
    managed_by  = "terraform"
  }
}

module "aks" {
  source = "../modules/aks"
  count  = var.runtime == "aks" ? 1 : 0

  cluster_name        = var.cluster_name
  resource_group_name = var.resource_group_name
  location            = var.location
  kubernetes_version  = "1.29"
  node_count          = 3
  vm_size             = "Standard_B4ms"
  autoscaling_enabled = true
  min_node_count      = 3
  max_node_count      = 8
  create_acr          = true
  acr_name            = "sagarslabstaging"
  tags                = var.tags
}

module "eks" {
  source = "../modules/eks"
  count  = var.runtime == "eks" ? 1 : 0

  cluster_name       = var.cluster_name
  aws_region         = var.aws_region
  kubernetes_version = "1.29"
  node_count         = 3
  min_node_count     = 3
  max_node_count     = 8
  instance_type      = "t3.large"
  create_ecr         = true
  ecr_repo_prefix    = "sagars-lab-staging"
  tags               = var.tags
}

output "cluster_name" {
  value = var.runtime == "aks" ? module.aks[0].cluster_name : module.eks[0].cluster_name
}
