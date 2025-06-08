output "minio_endpoint" {
  description = "MinIO API endpoint"
  value       = "minio-service.${var.namespace}.svc.cluster.local:9000"
}

output "minio_console_endpoint" {
  description = "MinIO Console endpoint"
  value       = "minio-service.${var.namespace}.svc.cluster.local:9001"
}

output "minio_root_user" {
  description = "MinIO root user"
  value       = var.minio_root_user
}

output "minio_root_password" {
  description = "MinIO root password"
  value       = var.minio_root_password
  sensitive   = true
}

output "namespace" {
  description = "Kubernetes namespace where MinIO is deployed"
  value       = var.namespace
}

output "service_name" {
  description = "MinIO Kubernetes service name"
  value       = kubernetes_service.minio_service.metadata[0].name
}

output "bucket_names" {
  description = "List of backup buckets created in MinIO"
  value = [
    "mongodb-backups",
    "influxdb-backups", 
    "scylladb-backups",
    "dgraph-backups",
    "redpanda-backups",
    "stream-backups"
  ]
}

output "storage_type" {
  description = "Type of storage being used (persistent or emptyDir)"
  value = var.use_persistent_storage ? "persistent" : "emptyDir"
}

output "storage_warning" {
  description = "Warning about data persistence"
  value = var.use_persistent_storage ? "Data will persist across pod restarts" : "WARNING: Using emptyDir - data will be lost when pod restarts!"
} 