variable "redpanda_name" {
  description = "Name for the Redpanda Helm release."
  type        = string
  default     = "redpanda"
}

variable "redpanda_repository" {
  description = "Helm chart repository for Redpanda."
  type        = string
  default     = "https://charts.redpanda.com/"
}

variable "redpanda_chart_name" {
  description = "Name of the Redpanda Helm chart."
  type        = string
  default     = "redpanda"
}

variable "redpanda_chart_version" {
  description = "Version of the Redpanda Helm chart."
  type        = string
  default     = "5.9.20" # Please check for the latest or desired version
}

variable "redpanda_namespace" {
  description = "Kubernetes namespace to deploy Redpanda into."
  type        = string
  default     = "redpanda"
}

variable "redpanda_console_enabled" {
  description = "Enable Redpanda Console."
  type        = bool
  default     = true
}

# Performance and scaling variables
variable "redpanda_replicas" {
  description = "Number of Redpanda broker replicas for high availability and throughput."
  type        = number
  default     = 1  # Single replica for memory-constrained environments
}

variable "redpanda_cpu_requests" {
  description = "CPU requests for Redpanda pods."
  type        = string
  default     = "1"  # Reduced for memory-constrained systems
}

variable "redpanda_memory_requests" {
  description = "Memory requests for Redpanda pods."
  type        = string
  default     = "2Gi"  # Reduced for memory-constrained systems
}

variable "redpanda_storage_size" {
  description = "Storage size for Redpanda persistent volumes."
  type        = string
  default     = "10Gi"  # Reduced for memory-constrained systems
}

variable "redpanda_default_partitions" {
  description = "Default number of partitions for new topics."
  type        = number
  default     = 3  # Fewer partitions for single broker
}

variable "redpanda_max_message_size" {
  description = "Maximum message size in bytes."
  type        = number
  default     = 10485760 # 10MB
}

variable "redpanda_fetch_max_bytes" {
  description = "Maximum bytes to fetch in a single request."
  type        = number
  default     = 52428800 # 50MB
}

variable "redpanda_compression_type" {
  description = "Default compression type for topics."
  type        = string
  default     = "snappy"
  validation {
    condition     = contains(["none", "gzip", "snappy", "lz4", "zstd"], var.redpanda_compression_type)
    error_message = "Compression type must be one of: none, gzip, snappy, lz4, zstd."
  }
}

variable "redpanda_log_retention_ms" {
  description = "Log retention time in milliseconds."
  type        = number
  default     = 604800000 # 7 days
}

# Example variables based on Ansible documentation provided
# You might need to adjust these or find corresponding Helm values

variable "redpanda_admin_api_port" {
  description = "Redpanda Admin API port (for reference, configure in Helm values)."
  type        = number
  default     = 9644
}

variable "redpanda_kafka_port" {
  description = "Redpanda Kafka API port (for reference, configure in Helm values)."
  type        = number
  default     = 9092
}

variable "redpanda_rpc_port" {
  description = "Redpanda RPC port (for reference, configure in Helm values)."
  type        = number
  default     = 33145
}

variable "redpanda_schema_registry_port" {
  description = "Redpanda Schema Registry port (for reference, configure in Helm values)."
  type        = number
  default     = 8081
}

variable "redpanda_rack_awareness_enabled" {
  description = "Enable rack awareness (for reference, configure in Helm values)."
  type        = bool
  default     = false # Defaulting to false as 'rack' is 'undefined' in Ansible defaults
}

variable "redpanda_tiered_storage_bucket_name" {
  description = "Name of the bucket for tiered storage (for reference, configure in Helm values). Set to enable."
  type        = string
  default     = "" # Empty string means disabled, as per 'Set bucket name to enable Tiered Storage.'
} 