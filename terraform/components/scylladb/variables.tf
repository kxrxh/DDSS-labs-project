variable "scylla_namespace" {
  description = "Kubernetes namespace for ScyllaDB clusters."
  type        = string
  default     = "scylla"
}

variable "scylla_operator_namespace" {
  description = "Kubernetes namespace for the ScyllaDB Operator."
  type        = string
  default     = "scylla-operator"
}

variable "scylla_operator_release_name" {
  description = "Helm release name for the ScyllaDB Operator."
  type        = string
  default     = "scylla-operator"
}

variable "scylla_operator_repository" {
  description = "Helm chart repository for the ScyllaDB Operator."
  type        = string
  default     = "https://scylla-operator-charts.storage.googleapis.com/stable"
}

variable "scylla_operator_chart_name" {
  description = "Helm chart name for the ScyllaDB Operator."
  type        = string
  default     = "scylla-operator"
}

variable "scylla_operator_chart_version" {
  description = "Helm chart version for the ScyllaDB Operator."
  type        = string
  default     = "1.16.2" # Please verify and update to your desired version
}

variable "scylla_operator_create_namespace" {
  description = "Whether to create the namespace for the ScyllaDB Operator if it doesn't exist."
  type        = bool
  default     = true
}

variable "scylla_agent_version" {
  description = "Version of the ScyllaDB Agent to use."
  type        = string
  default     = "3.2.1"
}

variable "scylla_cluster_name" {
  description = "Name of the ScyllaDB cluster to be created by the operator."
  type        = string
  default     = "scylla-cluster"
}

variable "scylla_cluster_datacenter_name" {
  description = "Name of the datacenter for the ScyllaDB cluster."
  type        = string
  default     = "dc1"
}

variable "scylla_cluster_rack_name" {
  description = "Name of the rack for the ScyllaDB cluster."
  type        = string
  default     = "rack1"
}

variable "scylla_cluster_members" {
  description = "Number of members (nodes) in the ScyllaDB cluster rack."
  type        = number
  default     = 1 # For local/dev setup; increase for production
}

variable "scylla_cluster_cpus" {
  description = "CPU request/limit for ScyllaDB nodes."
  type        = string
  default     = "1"
}

variable "scylla_cluster_memory" {
  description = "Memory request/limit for ScyllaDB nodes."
  type        = string
  default     = "2Gi" # Adjust based on your needs and available resources
} 