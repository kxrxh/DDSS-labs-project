output "cert_manager_ready" {
  description = "Indicator that cert-manager is ready"
  value       = null_resource.wait_for_cert_manager.id
} 