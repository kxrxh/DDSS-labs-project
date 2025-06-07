output "redpanda_service_endpoint" {
  description = "The endpoint for the Redpanda service."
  # This is a placeholder. You'll need to adjust this based on how the Helm chart exposes the service.
  # Example: value = helm_release.redpanda.status.load_balancer_ingress[0].hostname
  # Or for a ClusterIP service, you might output the service name and namespace.
  value       = "To be determined based on Helm chart output and service type"
}

output "redpanda_console_endpoint" {
  description = "The endpoint for the Redpanda Console if enabled."
  # This is also a placeholder and depends on how the console is exposed.
  value       = var.redpanda_console_enabled ? "To be determined based on Helm chart output and service type for console" : "Console not enabled"
} 