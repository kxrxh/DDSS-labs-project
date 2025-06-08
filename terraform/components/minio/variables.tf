variable "namespace" {
  type        = string
  default     = "minio"
  description = "The Kubernetes namespace to deploy MinIO into."
}

variable "replicas" {
  type        = number
  default     = 1
  description = "The number of MinIO replicas."
}

variable "storage_size" {
  type        = string
  default     = "50Gi"
  description = "The size of the persistent volume for MinIO data storage."
}

variable "storage_class" {
  type        = string
  default     = ""
  description = "The storage class to use for MinIO persistent volume claim."
}

variable "use_persistent_storage" {
  type        = bool
  default     = false
  description = "Whether to use persistent storage (PVC) or emptyDir for development."
}



variable "minio_image" {
  type        = string
  default     = "minio/minio:latest"
  description = "MinIO image to use"
}

variable "minio_root_user" {
  type        = string
  default     = "admin"
  description = "MinIO root user (admin username)"
}

variable "minio_root_password" {
  type        = string
  default     = "password123"
  description = "MinIO root password (admin password)"
  sensitive   = true
}

variable "memory_request" {
  type        = string
  default     = "256Mi"
  description = "Memory request for MinIO container"
}

variable "memory_limit" {
  type        = string
  default     = "512Mi"
  description = "Memory limit for MinIO container"
}

variable "cpu_request" {
  type        = string
  default     = "100m"
  description = "CPU request for MinIO container"
}

variable "cpu_limit" {
  type        = string
  default     = "250m"
  description = "CPU limit for MinIO container"
} 