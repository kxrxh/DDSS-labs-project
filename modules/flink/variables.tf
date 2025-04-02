variable "namespace" {
  description = "The Kubernetes namespace to deploy Flink Operator into."
  type        = string
  default     = "flink-operator"
}

variable "release_name" {
  description = "The Helm release name for Flink Operator."
  type        = string
  default     = "flink-operator"
}

variable "chart_version" {
  description = "The version of the Flink Kubernetes Operator Helm chart to use."
  type        = string
  default     = "1.11.0" // Updated to a version available at downloads.apache.org
}

variable "create_namespace" {
  description = "Whether the Helm chart should create the namespace."
  type        = bool
  default     = true
}

variable "watch_namespaces" {
  description = "List of namespaces the Flink operator should watch for FlinkDeployment resources."
  type        = list(string)
  default     = ["default", "flink"] # Add other namespaces where you'll deploy Flink jobs
}

variable "flink_operator_helm_config" {
  description = "Optional Helm configuration overrides for Flink operator. Keys are Helm value names (e.g., 'image.repository')."
  type        = map(any)
  default     = {}
}
