variable "release_name" {
  description = "Helm release name for InfluxDB."
  type        = string
  default     = "influxdb"
}

variable "namespace" {
  description = "Kubernetes namespace to deploy InfluxDB into."
  type        = string
}

variable "create_namespace" {
  description = "Whether the Helm chart should create the namespace if it doesn't exist."
  type        = bool
  default     = false
}

# variable "chart_version" {
#   description = "Version of the InfluxDB Helm chart to use."
#   type        = string
#   default     = "2.1.2" # Specify a recent, known working version
# }

variable "persistence_enabled" {
  description = "Enable persistence using PersistentVolumes."
  type        = bool
  default     = true
}

variable "persistence_size" {
  description = "Size of the persistent volume for InfluxDB data."
  type        = string
  default     = "8Gi"
}

variable "persistence_storage_class" {
  description = "StorageClass to use for persistence. If null, uses default."
  type        = string
  default     = null
} 