variable "cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
  default     = "sagars-cluster"
}

variable "resource_group_name" {
  description = "Azure resource group name"
  type        = string
  default     = "sagars-lab-rg"
}

variable "location" {
  description = "Azure region (overridden by resource group location)"
  type        = string
  default     = "eastus"
}

variable "kubernetes_version" {
  description = "Kubernetes version for the AKS cluster"
  type        = string
  default     = "1.29"
}

variable "node_count" {
  description = "Number of nodes in the default node pool"
  type        = number
  default     = 2
}

variable "vm_size" {
  description = "VM size for nodes"
  type        = string
  default     = "Standard_B2s"
}

variable "os_disk_size_gb" {
  description = "OS disk size in GB for nodes"
  type        = number
  default     = 30
}

variable "autoscaling_enabled" {
  description = "Enable cluster autoscaler"
  type        = bool
  default     = false
}

variable "min_node_count" {
  description = "Minimum node count when autoscaling is enabled"
  type        = number
  default     = 2
}

variable "max_node_count" {
  description = "Maximum node count when autoscaling is enabled"
  type        = number
  default     = 5
}

variable "create_acr" {
  description = "Create an Azure Container Registry"
  type        = bool
  default     = true
}

variable "acr_name" {
  description = "Name for the Azure Container Registry (must be globally unique)"
  type        = string
  default     = "sagarslab"
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
