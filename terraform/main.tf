provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "orbstack"
}

provider "helm" {
  kubernetes {
    config_path    = "~/.kube/config"
    config_context = "orbstack"
  }
}

module "mongodb" {
  source = "./components/mongodb"

  # You can override default MongoDB settings here if needed, for example:
  # mongodb_namespace    = "mongo-apps"
  # mongodb_auth_enabled = true
  # mongodb_chart_version = "15.5.0" # Pin to a specific chart version
}

module "redpanda" {
  source = "./components/redpanda"

  redpanda_name      = "dev-redpanda"
  redpanda_namespace = "dev-streaming"
  # You can override default Redpanda settings here if needed, for example:
  # redpanda_console_enabled = true

  # redpanda_tiered_storage_bucket_name = "my-redpanda-backup-bucket" # If you want to enable tiered storage
}

module "dgraph" {
  source = "./components/dgraph"

  # You can override default Dgraph settings here if needed, for example:
  # dgraph_namespace       = "dgraph-system"
  # dgraph_chart_version   = "24.1.1" # Pin to a specific chart version
  # dgraph_alpha_replicas  = 3
  # dgraph_zero_replicas   = 3 # For High Availability
}

module "scylladb" {
  source = "./components/scylladb"

  # Pass the cert-manager Helm release to ensure ScyllaDB operator waits for it
  cert_manager_helm_release = module.cert_manager.helm_release_cert_manager

  # You can override default ScyllaDB settings here if needed, for example:
  # scylla_namespace = "scylla-prod"
  # scylla_operator_chart_version = "1.12.1" # Pin to a specific chart version
  scylla_cluster_members = 1
  # scylla_cluster_cpus = "2"
  # scylla_cluster_memory = "4Gi"
}

module "influxdb" {
  source = "./components/influxdb"
  namespace = "influxdb"
  create_namespace = true

  # You can override default InfluxDB settings here if needed, for example:
  # namespace           = "influxdb-apps"
  # persistence_enabled = true
  # persistence_size    = "10Gi"
}

module "cert_manager" {
  source = "./components/cert-manager"
  # You can override default cert-manager variables here if needed, e.g.:
  # cert_manager_chart_version = "v1.13.3" # Pin to a specific version
}

module "minio" {
  source = "./components/minio"

  # Use emptyDir for fast development deployment (data won't persist across pod restarts)
  use_persistent_storage = false
  storage_size          = "10Gi"

  # You can override default MinIO settings here if needed, for example:
  # namespace           = "minio-backup"
  # use_persistent_storage = true  # Enable for production
  # storage_class       = "fast-ssd"
  # storage_size        = "100Gi"
  # minio_root_user     = "backup-admin"
  # minio_root_password = "secure-password-123"
  # replicas            = 2
}

locals {
  # Absolute path to the event-producer source directory, relative to the terraform root.
  event_producer_src_dir = abspath("${path.root}/../event-producer")
}

module "event_producer" {
  source = "./components/event-producer"

  event_producer_source_path = local.event_producer_src_dir
  kafka_brokers              = var.redpanda_bootstrap_servers_internal
  kafka_topic                = var.processing_input_kafka_topic # Ensure producer and processing consumer use the same topic
  dgraph_url                 = "dgraph-dgraph-alpha.dgraph.svc.cluster.local:9080"
}

locals {
  # Absolute path to the stream processor source directory
  stream_processor_src_dir = abspath("${path.root}/../stream")
}
module "stream_processor" {
  source = "./components/stream-processor"

  # Source path
  stream_processor_source_path = local.stream_processor_src_dir

  # Kafka configuration
  kafka_brokers    = var.redpanda_bootstrap_servers_internal
  kafka_topic      = var.processing_input_kafka_topic
  kafka_group_id   = "social-credit-processor"

  # MongoDB configuration
  mongo_uri        = "mongodb://mongodb-service.mongodb.svc.cluster.local:27017"
  mongo_database   = "social_rating"

  # InfluxDB configuration
  influxdb_url     = module.influxdb.influxdb_url
  influxdb_org     = module.influxdb.organization
  influxdb_token   = var.influxdb_token
  influxdb_bucket  = "events"
  influxdb_raw_events_bucket = "raw_events"
  influxdb_derived_metrics_bucket = "derived_metrics"

  # ScyllaDB configuration
  scylla_hosts     = "my-scylla-cluster-client.scylla.svc.cluster.local"
  scylla_keyspace  = "social_rating"

  # Dgraph configuration
  dgraph_hosts     = "dgraph-dgraph-alpha.dgraph.svc.cluster.local:9080"

  # MinIO configuration for backup
  minio_endpoint   = module.minio.minio_endpoint
  minio_access_key = module.minio.minio_root_user
  minio_secret_key = module.minio.minio_root_password
  minio_bucket_name = "stream-backups"

  # Processing configuration
  batch_size       = 500
  workers          = 8
  flush_interval   = "2s"
  replicas         = 2
  log_level        = "INFO"
}
