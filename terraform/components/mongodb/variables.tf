# modules/mongodb/variables.tf
variable "namespace" {
  type        = string
  default     = "mongodb"
  description = "The Kubernetes namespace to deploy MongoDB into."
}

variable "replicas" {
  type        = number
  default     = 1
  description = "The number of MongoDB replicas."
}

variable "storage_size" {
  type        = string
  default     = "10Gi"
  description = "The size of the persistent volume for each MongoDB instance."
}

variable "mongo_image" {
  type = string
  default = "mongo:latest"
  description = "mongo image to use"
}

variable "release_name" {
  description = "The Helm release name for MongoDB."
  type        = string
  default     = "mongodb" # Or whatever default makes sense
}