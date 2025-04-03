# Operator/Chart Variables
variable "operator_release_name" {
  description = "Helm release name for the ClickHouse deployment (using the main 'clickhouse' chart)."
  type        = string
  default     = "clickhouse"
}

variable "installation_namespace" {
  description = "Kubernetes namespace to deploy the ClickHouse chart into."
  type        = string
}

variable "operator_create_namespace" {
  description = "Whether the Helm chart should create the installation namespace."
  type        = bool
  default     = true # Usually fine to create the app namespace
}

variable "operator_chart_version" {
  description = "Version of the Altinity ClickHouse Helm chart ('clickhouse'). If null, latest will be used."
  type        = string
  default     = null # Use latest version
}