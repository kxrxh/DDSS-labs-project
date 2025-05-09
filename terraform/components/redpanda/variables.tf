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