variable "namespace" {
  description = "Kubernetes namespace for TiKV deployment"
  type        = string
}

variable "pd_replicas" {
  description = "Number of Placement Driver (PD) replicas"
  type        = number
}

variable "tikv_replicas" {
  description = "Number of TiKV replicas"
  type        = number
}

variable "pd_image" {
  description = "Docker image for PD"
  type        = string
}

variable "tikv_image" {
  description = "Docker image for TiKV"
  type        = string
}