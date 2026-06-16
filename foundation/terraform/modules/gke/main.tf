# GKE module — mirrors the aks/eks module shape (task 064).
# A regional GKE cluster with a separately-managed node pool and an optional
# Artifact Registry repo for app images. Verify against a real project once
# (same caveat as 038/039), record cost, then tear down.

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# Regional cluster. We remove the default node pool and attach our own so node
# config is managed explicitly (the recommended GKE pattern).
resource "google_container_cluster" "main" {
  name     = var.cluster_name
  project  = var.gcp_project
  location = var.gcp_region

  remove_default_node_pool = true
  initial_node_count       = 1

  # Pin to a version prefix when provided; otherwise track the REGULAR channel.
  min_master_version = var.kubernetes_version != "" ? var.kubernetes_version : null

  release_channel {
    channel = "REGULAR"
  }

  # Lab cluster: keep it simple. Networking uses the default VPC; for production
  # you would attach a dedicated VPC + secondary ranges.
  deletion_protection = false

  resource_labels = var.tags
}

resource "google_container_node_pool" "main" {
  name     = "${var.cluster_name}-pool"
  project  = var.gcp_project
  location = var.gcp_region
  cluster  = google_container_cluster.main.name

  node_count = var.node_count

  autoscaling {
    min_node_count = var.min_node_count
    max_node_count = var.max_node_count
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type = var.machine_type
    disk_size_gb = 50
    disk_type    = "pd-standard"

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]

    labels = var.tags
  }
}

# Artifact Registry (Docker) repo for app images, analogous to ACR/ECR.
resource "google_artifact_registry_repository" "apps" {
  count = var.create_gar ? 1 : 0

  project       = var.gcp_project
  location      = var.gcp_region
  repository_id = var.gar_repo
  format        = "DOCKER"
  description   = "App images for the Flightdeck simulator"
  labels        = var.tags
}
