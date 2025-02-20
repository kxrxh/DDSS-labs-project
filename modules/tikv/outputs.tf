# Outputs (optional, for debugging or reference)
output "pd_service_name" {
  value = kubernetes_service.pd_service.metadata[0].name
}

output "tikv_service_name" {
  value = kubernetes_service.tikv_service.metadata[0].name
}