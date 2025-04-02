output "release_name" {
  description = "The Helm release name of the Redpanda deployment."
  value       = helm_release.redpanda.name
}

output "namespace" {
  description = "The Kubernetes namespace Redpanda was deployed into."
  value       = helm_release.redpanda.namespace
}

output "external_kafka_advertised_listeners" {
  description = "The externally advertised Kafka listener addresses (if using NodePort/LoadBalancer). Needs parsing from service status."
  # Note: Getting the exact external address dynamically can be tricky with Helm/TF alone.
  # Often requires querying the Service object after deployment.
  # This output is a placeholder; manual inspection or a K8s data source might be needed.
  value       = "Check Kubernetes Service '${var.release_name}-external' in namespace '${var.namespace}' for NodePort/LoadBalancer IP and ports."
}

output "internal_kafka_bootstrap_servers" {
  description = "Internal Kafka bootstrap server addresses within the cluster."
  # This typically follows a predictable pattern based on release name and namespace.
  value       = "${var.release_name}.${var.namespace}.svc.cluster.local:9092"
}
