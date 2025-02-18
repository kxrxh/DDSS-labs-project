output "dgraph_zero_service_name" {
  value = kubernetes_service.dgraph_zero.metadata[0].name
}

output "dgraph_alpha_service_name" {
  value = kubernetes_service.dgraph_alpha.metadata[0].name
}
