variable "namespace" {
  description = "The Kubernetes namespace to deploy Redpanda into."
  type        = string
  default     = "redpanda"
}

variable "release_name" {
  description = "The Helm release name for Redpanda."
  type        = string
  default     = "redpanda"
}

variable "chart_version" {
  description = "The version of the Redpanda Helm chart to use."
  type        = string
  default     = "5.7.2" // Specify a recent, known-good version
}

variable "create_namespace" {
  description = "Whether the Helm chart should create the namespace."
  type        = bool
  default     = true
}

variable "storage_class_name" {
  description = "The storage class to use for Redpanda data persistence. Should exist in the cluster."
  type        = string
  default     = "standard" // Adjust if you use a different default or custom StorageClass
}

variable "replicas" {
  description = "Number of Redpanda broker replicas."
  type        = number
  default     = 1 // For dev/testing; increase for HA
}
