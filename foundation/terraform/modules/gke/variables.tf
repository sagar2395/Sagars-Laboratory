variable "cluster_name" {
  description = "Name of the GKE cluster"
  type        = string
  default     = "sagars-cluster"
}

variable "gcp_project" {
  description = "GCP project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region for the cluster (regional cluster)"
  type        = string
  default     = "us-central1"
}

variable "kubernetes_version" {
  description = "Kubernetes release channel version prefix (e.g. 1.29). Empty uses the channel default."
  type        = string
  default     = ""
}

variable "node_count" {
  description = "Number of nodes per zone in the default node pool"
  type        = number
  default     = 1
}

variable "min_node_count" {
  description = "Minimum nodes per zone (autoscaling)"
  type        = number
  default     = 1
}

variable "max_node_count" {
  description = "Maximum nodes per zone (autoscaling)"
  type        = number
  default     = 3
}

variable "machine_type" {
  description = "Machine type for worker nodes"
  type        = string
  default     = "e2-standard-2"
}

variable "create_gar" {
  description = "Create an Artifact Registry repository for app images"
  type        = bool
  default     = true
}

variable "gar_repo" {
  description = "Artifact Registry repository name"
  type        = string
  default     = "sagars-lab"
}

variable "tags" {
  description = "Resource labels (GCP labels must be lowercase)"
  type        = map(string)
  default = {
    project     = "sagars-laboratory"
    environment = "dev"
    managed_by  = "terraform"
  }
}
