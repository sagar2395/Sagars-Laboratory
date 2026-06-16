output "cluster_name" {
  description = "Name of the GKE cluster"
  value       = google_container_cluster.main.name
}

output "cluster_endpoint" {
  description = "GKE cluster API endpoint"
  value       = google_container_cluster.main.endpoint
  sensitive   = true
}

output "cluster_location" {
  description = "Region of the GKE cluster"
  value       = google_container_cluster.main.location
}

output "gar_repository_url" {
  description = "Artifact Registry Docker repository URL (empty if not created)"
  value = var.create_gar ? format(
    "%s-docker.pkg.dev/%s/%s",
    var.gcp_region, var.gcp_project, var.gar_repo
  ) : ""
}
