output "helm_release_cert_manager" {
  description = "The cert-manager Helm release resource."
  value       = helm_release.cert_manager
} 