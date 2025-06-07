variable "event_producer_namespace" {
  description = "Kubernetes namespace for the event producer"
  type        = string
  default     = "event-producer"
}

variable "create_namespace" {
  description = "Whether to create the namespace"
  type        = bool
  default     = true
}

variable "event_producer_source_path" {
  description = "Path to the event producer source code"
  type        = string
  default     = "../../event-producer"
}

variable "event_producer_image_name" {
  description = "Name of the event producer Docker image"
  type        = string
  default     = "event-producer"
}

variable "event_producer_image_tag" {
  description = "Tag for the event producer Docker image"
  type        = string
  default     = "latest"
}

variable "kafka_brokers" {
  description = "Comma-separated list of Kafka brokers"
  type        = string
  default     = "dev-redpanda.dev-streaming.svc.cluster.local:9093"
}

variable "kafka_topic" {
  description = "Kafka topic for events"
  type        = string
  default     = "citizen_events"
}

variable "num_citizens" {
  description = "Number of base citizens to create"
  type        = number
  default     = 100
}

variable "send_interval" {
  description = "Interval between event generation"
  type        = string
  default     = "2s"
}

variable "replicas" {
  description = "Number of event producer replicas"
  type        = number
  default     = 1
}

variable "mongo_uri" {
  description = "MongoDB connection URI"
  type        = string
  default     = "mongodb://mongodb-service.mongodb.svc.cluster.local:27017"
}

variable "dgraph_url" {
  description = "Dgraph connection URL"
  type        = string
  default     = "dgraph-dgraph-alpha.dgraph.svc.cluster.local:9080"
} 