terraform {
  required_version = ">= 1.5.0"

  # Uncomment to use remote state:
  # backend "azurerm" {
  #   resource_group_name  = "terraform-state-rg"
  #   storage_account_name = "sagarslabstate"
  #   container_name       = "tfstate"
  #   key                  = "dev.terraform.tfstate"
  # }
  # backend "s3" {
  #   bucket = "sagars-lab-tfstate"
  #   key    = "dev/terraform.tfstate"
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
  default     = "sagars-cluster"
}

# AKS-specific variables
variable "resource_group_name" {
  description = "Azure resource group name (AKS only)"
  type        = string
  default     = "sagars-lab-rg"
}

variable "location" {
  description = "Azure region (AKS only)"
  type        = string
  default     = "eastus"
}

# EKS-specific variables
variable "aws_region" {
  description = "AWS region (EKS only)"
  type        = string
  default     = "us-east-1"
}

# Shared variables
variable "node_count" {
  description = "Number of worker nodes"
  type        = number
  default     = 2
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    project     = "sagars-laboratory"
    environment = "dev"
    managed_by  = "terraform"
  }
}

# AKS module
module "aks" {
  source = "../modules/aks"
  count  = var.runtime == "aks" ? 1 : 0

  cluster_name        = var.cluster_name
  resource_group_name = var.resource_group_name
  location            = var.location
  kubernetes_version  = var.kubernetes_version
  node_count          = var.node_count
  vm_size             = "Standard_B2s"
  create_acr          = true
  tags                = var.tags
}

# EKS module
module "eks" {
  source = "../modules/eks"
  count  = var.runtime == "eks" ? 1 : 0

  cluster_name       = var.cluster_name
  aws_region         = var.aws_region
  kubernetes_version = var.kubernetes_version
  node_count         = var.node_count
  instance_type      = "t3.medium"
  create_ecr         = true
  tags               = var.tags
}

# Outputs — runtime-agnostic
output "cluster_name" {
  value = var.runtime == "aks" ? module.aks[0].cluster_name : module.eks[0].cluster_name
}

output "registry_url" {
  value = var.runtime == "aks" ? module.aks[0].acr_login_server : (
    length(module.eks) > 0 ? join(",", [for k, v in module.eks[0].ecr_repository_urls : v]) : ""
  )
}
