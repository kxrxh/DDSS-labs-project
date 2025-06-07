resource "null_resource" "build_stream_processor_image" {
  # Trigger this resource to rebuild if the source code of the stream processor changes.
  triggers = {
    go_mod_hash     = sha1(file("${var.stream_processor_source_path}/go.mod"))
    go_sum_hash     = sha1(file("${var.stream_processor_source_path}/go.sum"))
    dockerfile_hash = sha1(file("${var.stream_processor_source_path}/Dockerfile"))
    src_hash = sha1(join("", [
      for f in fileset("${var.stream_processor_source_path}/", "**/*.go") :
      filemd5("${var.stream_processor_source_path}/${f}")
    ]))
  }

  provisioner "local-exec" {
    # Build the Docker image.
    command = "docker build -t ${var.stream_processor_image_name}:${var.stream_processor_image_tag} ${var.stream_processor_source_path}"
  }
}

resource "kubernetes_namespace" "stream_processor_ns" {
  count = var.create_namespace ? 1 : 0
  metadata {
    name = var.stream_processor_namespace
  }
}

# ConfigMap for stream processor configuration
resource "kubernetes_config_map_v1" "stream_processor_config" {
  metadata {
    name      = "stream-processor-config"
    namespace = var.stream_processor_namespace
  }

  data = {
    "KAFKA_BROKERS"     = var.kafka_brokers
    "KAFKA_TOPIC"       = var.kafka_topic
    "KAFKA_GROUP_ID"    = var.kafka_group_id
    "MONGO_URI"         = var.mongo_uri
    "MONGO_DATABASE"    = var.mongo_database
    "INFLUX_URL"        = var.influxdb_url
    "INFLUX_ORG"        = var.influxdb_org
    "INFLUX_BUCKET"     = var.influxdb_bucket
    "INFLUX_RAW_EVENTS_BUCKET" = var.influxdb_raw_events_bucket
    "INFLUX_DERIVED_METRICS_BUCKET" = var.influxdb_derived_metrics_bucket
    "SCYLLA_HOSTS"      = var.scylla_hosts
    "SCYLLA_KEYSPACE"   = var.scylla_keyspace
    "DGRAPH_HOSTS"      = var.dgraph_hosts
    "BATCH_SIZE"        = tostring(var.batch_size)
    "WORKERS"           = tostring(var.workers)
    "FLUSH_INTERVAL"    = var.flush_interval
    "LOG_LEVEL"         = var.log_level
  }

  depends_on = [kubernetes_namespace.stream_processor_ns]
}

# Secret for sensitive configuration (InfluxDB token)
resource "kubernetes_secret_v1" "stream_processor_secrets" {
  metadata {
    name      = "stream-processor-secrets"
    namespace = var.stream_processor_namespace
  }

  data = {
    "INFLUX_TOKEN" = var.influxdb_token
  }

  type = "Opaque"

  depends_on = [kubernetes_namespace.stream_processor_ns]
}

# Deployment for the stream processor
resource "kubernetes_deployment_v1" "stream_processor_deployment" {
  metadata {
    name      = "stream-processor"
    namespace = var.stream_processor_namespace
    labels = {
      app     = "stream-processor"
      version = var.stream_processor_image_tag
    }
  }

  spec {
    replicas = var.replicas
    
    selector {
      match_labels = {
        app = "stream-processor"
      }
    }

    template {
      metadata {
        labels = {
          app     = "stream-processor"
          version = var.stream_processor_image_tag
        }
      }

      spec {
        container {
          name  = "stream-processor"
          image = "${var.stream_processor_image_name}:${var.stream_processor_image_tag}"
          image_pull_policy = "IfNotPresent" # Important for locally built images

          # Environment variables from ConfigMap
          env_from {
            config_map_ref {
              name = kubernetes_config_map_v1.stream_processor_config.metadata[0].name
            }
          }

          # Environment variables from Secret
          env_from {
            secret_ref {
              name = kubernetes_secret_v1.stream_processor_secrets.metadata[0].name
            }
          }

          # Resource limits and requests - increased for higher throughput
          resources {
            limits = {
              memory = "2Gi"
              cpu    = "2000m"
            }
            requests = {
              memory = "1Gi"
              cpu    = "1000m"
            }
          }

          # Health checks
          liveness_probe {
            http_get {
              path = "/health"
              port = 8080
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          readiness_probe {
            http_get {
              path = "/ready"
              port = 8080
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 3
            failure_threshold     = 3
          }

          # Ports
          port {
            container_port = 8080
            name          = "http"
            protocol      = "TCP"
          }
        }

        # Pod restart policy
        restart_policy = "Always"

        # Security context
        security_context {
          run_as_non_root = true
          run_as_user     = 1000
          run_as_group    = 3000
          fs_group        = 2000
        }
      }
    }

    # Deployment strategy
    strategy {
      type = "RollingUpdate"
      rolling_update {
        max_unavailable = "25%"
        max_surge       = "25%"
      }
    }
  }

  depends_on = [
    null_resource.build_stream_processor_image,
    kubernetes_config_map_v1.stream_processor_config,
    kubernetes_secret_v1.stream_processor_secrets,
    var.mongodb_init_job,
  ]
}

# Service for the stream processor (for health checks and monitoring)
resource "kubernetes_service_v1" "stream_processor_service" {
  metadata {
    name      = "stream-processor-service"
    namespace = var.stream_processor_namespace
    labels = {
      app = "stream-processor"
    }
  }

  spec {
    selector = {
      app = "stream-processor"
    }

    port {
      name        = "http"
      port        = 8080
      target_port = 8080
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }

  depends_on = [kubernetes_deployment_v1.stream_processor_deployment]
} 