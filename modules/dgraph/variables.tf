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
