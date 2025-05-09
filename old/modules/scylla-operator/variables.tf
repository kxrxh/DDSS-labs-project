variable "namespace" {
  description = "The Kubernetes namespace to deploy ScyllaDB Operator into."
  type        = string
  default     = "scylla-operator"
}

variable "release_name" {
  description = "The Helm release name for ScyllaDB Operator."
  type        = string
  default     = "scylla-operator"
}

variable "create_namespace" {
  description = "Whether to create the Kubernetes namespace if it doesn't exist."
  type        = bool
  default     = true
}

variable "chart_version" {
  description = "The version of the ScyllaDB Operator Helm chart to use."
  type        = string
  default     = "1.11.0" # A recent stable version
}
