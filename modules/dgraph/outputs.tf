output "dgraph_zero_service_name" {
  value = kubernetes_service.dgraph_zero.metadata[0].name
}

output "dgraph_alpha_service_name" {
  value = kubernetes_service.dgraph_alpha.metadata[0].name
}

output "dgraph_ratel_service_name" {
  value = kubernetes_service.dgraph_ratel.metadata[0].name
}

output "dgraph_dashboard_url" {
  value = "http://${kubernetes_service.dgraph_ratel.metadata[0].name}.${var.namespace}:8000"
  description = "URL to access the Dgraph Ratel dashboard (accessible within the cluster)"
}
