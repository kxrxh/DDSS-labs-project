# InfluxDB service information
output "influxdb_url" {
  description = "InfluxDB service URL for internal cluster access"
  value       = "http://${var.release_name}-influxdb2.${var.namespace}.svc.cluster.local:8086"
}

output "influxdb_external_url" {
  description = "InfluxDB external URL (if exposed)"
  value       = "http://localhost:8086"
}

# Authentication and tokens
output "influxdb_admin_secret" {
  description = "Name of the Kubernetes secret containing InfluxDB tokens"
  value       = "influxdb-tokens"
}

output "organization" {
  description = "InfluxDB organization name"
  value       = var.organization
}

output "admin_username" {
  description = "InfluxDB admin username"
  value       = var.admin_username
}

# Bucket information
output "buckets_created" {
  description = "List of buckets that will be created during initialization"
  value = [
    "raw_events",
    "derived_metrics", 
    "system_metrics",
    var.default_bucket
  ]
}

output "raw_events_bucket" {
  description = "Name of the raw events bucket"
  value       = "raw_events"
}

output "derived_metrics_bucket" {
  description = "Name of the derived metrics bucket"
  value       = "derived_metrics"
}

output "system_metrics_bucket" {
  description = "Name of the system metrics bucket"
  value       = "system_metrics"
}

# Service discovery information
output "service_name" {
  description = "InfluxDB Kubernetes service name"
  value       = "${var.release_name}-influxdb2"
}

output "service_port" {
  description = "InfluxDB service port"
  value       = 8086
}

output "namespace" {
  description = "Kubernetes namespace where InfluxDB is deployed"
  value       = var.namespace
}

# Job status
output "init_job_name" {
  description = "Name of the InfluxDB initialization job"
  value       = "${var.release_name}-init"
}

# Connection configuration for applications
output "connection_config" {
  description = "Connection configuration for applications"
  value = {
    url         = "http://${var.release_name}-influxdb2.${var.namespace}.svc.cluster.local:8086"
    org         = var.organization
    token_secret = "influxdb-tokens"
    buckets = {
      raw_events      = "raw_events"
      derived_metrics = "derived_metrics"
      system_metrics  = "system_metrics"
      default        = var.default_bucket
    }
  }
} 