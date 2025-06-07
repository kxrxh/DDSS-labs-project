# modules/mongodb/outputs.tf

output "headless_service_name" {
  value = kubernetes_service.mongodb_headless.metadata[0].name
}

output "statefulset_name" {
  value = kubernetes_stateful_set.mongodb.metadata[0].name
}

output "mongo_express_service_name" {
  value = kubernetes_service.mongo_express.metadata[0].name
}

output "initialization_job" {
  description = "MongoDB initialization job resource"
  value       = kubernetes_job_v1.mongodb_init_job
}
