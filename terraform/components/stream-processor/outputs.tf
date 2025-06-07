output "namespace" {
  description = "The Kubernetes namespace where the stream processor is deployed"
  value       = var.stream_processor_namespace
}

output "deployment_name" {
  description = "The name of the stream processor deployment"
  value       = kubernetes_deployment_v1.stream_processor_deployment.metadata[0].name
}

output "service_name" {
  description = "The name of the stream processor service"
  value       = kubernetes_service_v1.stream_processor_service.metadata[0].name
}

output "service_endpoint" {
  description = "The internal service endpoint for the stream processor"
  value       = "${kubernetes_service_v1.stream_processor_service.metadata[0].name}.${var.stream_processor_namespace}.svc.cluster.local:8080"
}

output "image_name" {
  description = "The Docker image name used for the stream processor"
  value       = "${var.stream_processor_image_name}:${var.stream_processor_image_tag}"
} 