output "release_name" {
  description = "The Helm release name of the Flink Operator deployment."
  value       = helm_release.flink_operator.name
}

output "namespace" {
  description = "The Kubernetes namespace Flink Operator was deployed into."
  value       = helm_release.flink_operator.namespace
}

output "watched_namespaces" {
  description = "Namespaces monitored by the Flink Operator."
  value       = var.watch_namespaces
}
