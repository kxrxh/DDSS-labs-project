variable "namespace" {
  description = "The Kubernetes namespace to deploy ScyllaDB into."
  type        = string
  default     = "scylladb"
}

variable "release_name" {
  description = "The Helm release name for ScyllaDB."
  type        = string
  default     = "scylladb"
}

variable "create_namespace" {
  description = "Whether to create the Kubernetes namespace if it doesn't exist."
  type        = bool
  default     = true
}

variable "chart_version" {
  description = "The version of the ScyllaDB Helm chart to use."
  type        = string
  default     = "1.9.0" # Specify a recent, stable version
}

variable "persistence_storage_class" {
  description = "The storage class to use for ScyllaDB persistence."
  type        = string
  default     = "standard" # Or your default SC, e.g., local-path
}

variable "persistence_size" {
  description = "The size of the persistent volume for ScyllaDB."
  type        = string
  default     = "10Gi" # Keep it reasonable for local dev
} 