 variable "cert_manager_version" {
  description = "The version of cert-manager to install"
  type        = string
  default     = "1.14.3"
}

variable "timeout" {
  description = "Timeout in seconds for waiting on deployments"
  type        = number
  default     = 300
}