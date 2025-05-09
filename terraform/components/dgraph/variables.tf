variable "dgraph_name" {
  description = "Name for the Dgraph Helm release."
  type        = string
  default     = "dgraph"
}

variable "dgraph_repository" {
  description = "Helm chart repository for Dgraph."
  type        = string
  default     = "https://charts.dgraph.io"
}

variable "dgraph_chart_name" {
  description = "Name of the Dgraph Helm chart."
  type        = string
  default     = "dgraph"
}

variable "dgraph_chart_version" {
  description = "Version of the Dgraph Helm chart."
  type        = string
  default     = "24.1.1"
}

variable "dgraph_namespace" {
  description = "Kubernetes namespace to deploy Dgraph into."
  type        = string
  default     = "dgraph"
}

# Add other variables as needed, for example:
# variable "dgraph_alpha_replicas" {
#   description = "Number of Dgraph Alpha replicas."
#   type        = number
#   default     = 1
# }

# variable "dgraph_zero_replicas" {
#   description = "Number of Dgraph Zero replicas."
#   type        = number
#   default     = 1 # For HA, consider 3
# }

# variable "dgraph_alpha_persistence_enabled" {
#   description = "Enable persistence for Dgraph Alpha nodes."
#   type        = bool
#   default     = true
# }

# variable "dgraph_alpha_persistence_size" {
#   description = "Persistence volume size for Dgraph Alpha nodes."
#   type        = string
#   default     = "10Gi"
# }

# variable "dgraph_zero_persistence_enabled" {
#   description = "Enable persistence for Dgraph Zero nodes."
#   type        = bool
#   default     = true
# }

# variable "dgraph_zero_persistence_size" {
#   description = "Persistence volume size for Dgraph Zero nodes."
#   type        = string
#   default     = "1Gi"
# } 