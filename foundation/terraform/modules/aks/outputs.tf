output "cluster_name" {
  description = "Name of the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.name
}

output "cluster_id" {
  description = "AKS cluster resource ID"
  value       = azurerm_kubernetes_cluster.main.id
}

output "kube_config_raw" {
  description = "Raw kubeconfig for the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.kube_config_raw
  sensitive   = true
}

output "cluster_fqdn" {
  description = "FQDN of the AKS cluster API server"
  value       = azurerm_kubernetes_cluster.main.fqdn
}

output "kubelet_identity" {
  description = "Kubelet managed identity object ID"
  value       = azurerm_kubernetes_cluster.main.kubelet_identity[0].object_id
}

output "node_resource_group" {
  description = "Auto-generated resource group for AKS node resources"
  value       = azurerm_kubernetes_cluster.main.node_resource_group
}

output "acr_login_server" {
  description = "ACR login server URL"
  value       = var.create_acr ? azurerm_container_registry.main[0].login_server : ""
}

output "acr_id" {
  description = "ACR resource ID"
  value       = var.create_acr ? azurerm_container_registry.main[0].id : ""
}

output "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID"
  value       = azurerm_log_analytics_workspace.main.id
}
