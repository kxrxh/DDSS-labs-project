variable "namespace" {
  type        = string
  description = "The Kubernetes namespace to deploy Dgraph into."
  default     = "dgraph"
}

variable "zero_replicas" {
  type        = number
  description = "Number of replicas for Dgraph Zero."
  default     = 1
}

variable "alpha_replicas" {
  type        = number
  description = "Number of replicas for Dgraph Alpha."
  default     = 1
}

variable "zero_storage_size" {
  type        = string
  description = "Storage size for Dgraph Zero."
  default     = "10Gi"
}

variable "alpha_storage_size" {
  type        = string
  description = "Storage size for Dgraph Alpha."
  default     = "10Gi"
}

variable "dgraph_zero_image" {
  type = string
  description = "The docker image for dgraph zero"
  default = "dgraph/dgraph:latest"
}

variable "dgraph_alpha_image" {
  type = string
  description = "The docker image for dgraph alpha"
  default = "dgraph/dgraph:latest"
}

variable "dgraph_ratel_image" {
  type        = string
  description = "The docker image for Dgraph Ratel dashboard"
  default     = "dgraph/ratel:latest"
}

variable "ratel_replicas" {
  type        = number
  description = "Number of replicas for Dgraph Ratel dashboard"
  default     = 1
}

variable "ratel_service_type" {
  type        = string
  description = "Service type for exposing Ratel dashboard"
  default     = "ClusterIP"
}

variable "ratel_cpu_request" {
  type        = string
  description = "CPU request for Ratel dashboard"
  default     = "100m"
}

variable "ratel_memory_request" {
  type        = string
  description = "Memory request for Ratel dashboard"
  default     = "128Mi"
}

variable "ratel_cpu_limit" {
  type        = string
  description = "CPU limit for Ratel dashboard"
  default     = "200m"
}

variable "ratel_memory_limit" {
  type        = string
  description = "Memory limit for Ratel dashboard"
  default     = "256Mi"
}
