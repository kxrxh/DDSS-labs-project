variable "redpanda_bootstrap_servers_internal" {
  description = "Internal Redpanda bootstrap servers for Kubernetes cluster communication"
  type        = string
  default     = "dev-redpanda.dev-streaming.svc.cluster.local:9093"
}

variable "processing_input_kafka_topic" {
  description = "Kafka topic for processing input events"
  type        = string
  default     = "citizen_events"
}

variable "influxdb_token" {
  description = "InfluxDB authentication token"
  type        = string
  sensitive   = true
  default     = "SJlclSgz9_7kcda4525xgHreow5f64j7GWZtzVM79LQ35KAtjy3y306Djomc5hgD25WvERI0qqvsaGZ0_VAx7w=="
} 