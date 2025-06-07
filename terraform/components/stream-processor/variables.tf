variable "stream_processor_namespace" {
  description = "Kubernetes namespace for the stream processor"
  type        = string
  default     = "stream-processor"
}

variable "create_namespace" {
  description = "Whether to create the namespace"
  type        = bool
  default     = true
}

variable "stream_processor_source_path" {
  description = "Path to the stream processor source code"
  type        = string
  default     = "../../stream"
}

variable "stream_processor_image_name" {
  description = "Name of the stream processor Docker image"
  type        = string
  default     = "stream-processor"
}

variable "stream_processor_image_tag" {
  description = "Tag for the stream processor Docker image"
  type        = string
  default     = "latest"
}

variable "kafka_brokers" {
  description = "Comma-separated list of Kafka brokers"
  type        = string
  default     = "dev-redpanda.dev-streaming.svc.cluster.local:9093"
}

variable "kafka_topic" {
  description = "Kafka topic for input events"
  type        = string
  default     = "citizen_events"
}

variable "kafka_group_id" {
  description = "Kafka consumer group ID"
  type        = string
  default     = "social-credit-processor"
}

variable "mongo_uri" {
  description = "MongoDB connection URI"
  type        = string
  default     = "mongodb://mongodb-service.mongodb.svc.cluster.local:27017"
}

variable "mongo_database" {
  description = "MongoDB database name"
  type        = string
  default     = "social_rating"
}

variable "influxdb_url" {
  description = "InfluxDB connection URL"
  type        = string
  default     = "http://influxdb.influxdb.svc.cluster.local:8086"
}

variable "influxdb_token" {
  description = "InfluxDB authentication token"
  type        = string
  sensitive   = true
  default     = "3vndX0dIQ_RLAzIc3k0eo7WRTrW0sz-MbnNacN9DcZXpWdE6QsdPtL20l2y_KOhxeoBiwFhFgeJ8dbLvPvQrkw=="
}

variable "influxdb_org" {
  description = "InfluxDB organization"
  type        = string
  default     = "social-credit"
}

variable "influxdb_bucket" {
  description = "InfluxDB bucket name"
  type        = string
  default     = "events"
}

variable "influxdb_raw_events_bucket" {
  description = "InfluxDB bucket name for raw events"
  type        = string
  default     = "raw_events"
}

variable "influxdb_derived_metrics_bucket" {
  description = "InfluxDB bucket name for derived metrics"
  type        = string
  default     = "derived_metrics"
}

variable "scylla_hosts" {
  description = "ScyllaDB hosts (comma-separated)"
  type        = string
  default     = "my-scylla-cluster-client.scylla.svc.cluster.local"
}

variable "scylla_keyspace" {
  description = "ScyllaDB keyspace"
  type        = string
  default     = "social_rating"
}

variable "dgraph_hosts" {
  description = "Dgraph hosts (comma-separated)"
  type        = string
  default     = "dgraph-dgraph-alpha.dgraph.svc.cluster.local:9080"
}

variable "batch_size" {
  description = "Batch size for processing events"
  type        = number
  default     = 500
}

variable "workers" {
  description = "Number of worker goroutines"
  type        = number
  default     = 8
}

variable "flush_interval" {
  description = "Flush interval for batched operations"
  type        = string
  default     = "2s"
}

variable "replicas" {
  description = "Number of stream processor replicas"
  type        = number
  default     = 3
}

variable "log_level" {
  description = "Log level for the application"
  type        = string
  default     = "INFO"
}

variable "mongodb_init_job" {
  description = "MongoDB initialization job dependency"
  type        = any
  default     = null
} 