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

# Admin user configuration
variable "admin_username" {
  description = "InfluxDB admin username"
  type        = string
  default     = "admin"
}

variable "admin_password" {
  description = "InfluxDB admin password"
  type        = string
  default     = "admin123"
  sensitive   = true
}

variable "admin_token" {
  description = "InfluxDB admin token (if already exists)"
  type        = string
  default     = ""
  sensitive   = true
}

# Organization and bucket configuration
variable "organization" {
  description = "InfluxDB organization name"
  type        = string
  default     = "social-rating"
}

variable "default_bucket" {
  description = "Default bucket name created during setup"
  type        = string
  default     = "events"
}

variable "default_retention_seconds" {
  description = "Default retention period in seconds for the initial bucket"
  type        = number
  default     = 2592000 # 30 days
}

# Additional configuration
variable "enable_cleanup_job" {
  description = "Enable cleanup job for development purposes"
  type        = bool
  default     = false
} 