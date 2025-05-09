output "dgraph_alpha_service_endpoint" {
  description = "The endpoint for the Dgraph Alpha service (HTTP)."
  # This will depend on how the Dgraph Helm chart exposes the Alpha service (e.g., LoadBalancer, NodePort, ClusterIP).
  # For a LoadBalancer, it might be: helm_release.dgraph.status.load_balancer_ingress[0].hostname or .ip
  # For ClusterIP, it would be the internal service DNS: <release-name>-alpha.<namespace>.svc.cluster.local:8080
  value       = "To be determined based on Helm chart output and service type for Dgraph Alpha (HTTP: typically port 8080)"
}

output "dgraph_alpha_grpc_service_endpoint" {
  description = "The endpoint for the Dgraph Alpha gRPC service."
  # Similar to the HTTP endpoint, depends on service exposure.
  # For ClusterIP, it would be the internal service DNS: <release-name>-alpha.<namespace>.svc.cluster.local:9080
  value       = "To be determined based on Helm chart output and service type for Dgraph Alpha (gRPC: typically port 9080)"
}

output "dgraph_zero_service_endpoint" {
  description = "The endpoint for the Dgraph Zero service."
  # Dgraph Zero nodes also expose services, typically ClusterIP.
  # Example: <release-name>-zero.<namespace>.svc.cluster.local:5080 (gRPC) and :6080 (HTTP)
  value       = "To be determined based on Helm chart output and service type for Dgraph Zero"
} 