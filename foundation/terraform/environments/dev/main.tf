terraform {
  required_version = ">= 1.5.0"
  # Backend configuration lives in backend.tf (not checked in — see backend.tf.example).
  # Without a backend.tf the state is stored locally in terraform.tfstate,
  # which is fine for solo k3d experiments but not for shared/CI use.

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.80"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.30"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# Provider configuration. Both providers are declared because a single root
# handles either runtime; only the one matching var.runtime evaluates resources
# (the other module has count = 0, so its provider is never authenticated).
# azurerm REQUIRES a features {} block or `terraform apply` fails outright.
provider "azurerm" {
  features {}
}

provider "aws" {
  region = var.aws_region
}

# The google provider authenticates from Application Default Credentials
# (gcloud auth application-default login). project/region are taken from vars so
# the provider is inert when var.runtime != "gke" (the module has count = 0).
provider "google" {
  project = var.gcp_project != "" ? var.gcp_project : null
  region  = var.gcp_region
}

variable "runtime" {
  description = "Which cloud runtime to provision: aks, eks, or gke"
  type        = string

  validation {
    condition     = contains(["aks", "eks", "gke"], var.runtime)
    error_message = "runtime must be 'aks', 'eks', or 'gke'"
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

# GKE-specific variables
variable "gcp_project" {
  description = "GCP project ID (GKE only)"
  type        = string
  default     = ""
}

variable "gcp_region" {
  description = "GCP region (GKE only)"
  type        = string
  default     = "us-central1"
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

# GKE module
module "gke" {
  source = "../modules/gke"
  count  = var.runtime == "gke" ? 1 : 0

  cluster_name = var.cluster_name
  gcp_project  = var.gcp_project
  gcp_region   = var.gcp_region
  node_count   = var.node_count
  machine_type = "e2-standard-2"
  create_gar   = true
  tags         = var.tags
}

# Outputs — runtime-agnostic
output "cluster_name" {
  value = (
    var.runtime == "aks" ? module.aks[0].cluster_name :
    var.runtime == "eks" ? module.eks[0].cluster_name :
    module.gke[0].cluster_name
  )
}

output "registry_url" {
  value = (
    var.runtime == "aks" ? module.aks[0].acr_login_server :
    var.runtime == "eks" ? (length(module.eks) > 0 ? join(",", [for k, v in module.eks[0].ecr_repository_urls : v]) : "") :
    module.gke[0].gar_repository_url
  )
}
