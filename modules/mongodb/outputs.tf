# modules/mongodb/outputs.tf

output "headless_service_name" {
  value = kubernetes_service.mongodb_headless.metadata[0].name
}

output "statefulset_name" {
  value = kubernetes_stateful_set.mongodb.metadata[0].name
}
