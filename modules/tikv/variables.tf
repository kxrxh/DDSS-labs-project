variable "namespace" {
  description = "Kubernetes namespace for TiKV deployment"
  type        = string
  default     = "tikv-dev"
}

variable "pd_replicas" {
  description = "Number of Placement Driver (PD) replicas"
  type        = number
  default     = 1
}

variable "tikv_replicas" {
  description = "Number of TiKV replicas"
  type        = number
  default     = 1
}

variable "pd_storage_size" {
  description = "Storage size for PD nodes"
  type        = string
  default     = "2Gi"
}

variable "tikv_storage_size" {
  description = "Storage size for TiKV nodes"
  type        = string
  default     = "5Gi"
}

variable "pd_image" {
  description = "Docker image for PD"
  type        = string
  default     = "pingcap/pd:latest"
}

variable "tikv_image" {
  description = "Docker image for TiKV"
  type        = string
  default     = "pingcap/tikv:latest"
}