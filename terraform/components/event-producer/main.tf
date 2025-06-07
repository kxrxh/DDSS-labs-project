resource "null_resource" "build_event_producer_image" {
  # Trigger this resource to rebuild if the source code of the event producer changes.
  # This is a basic trigger; for more complex scenarios, consider externalizing image building.
  triggers = {
    # Hash of all .go files, go.mod, and Dockerfile in the event-producer directory.
    # Assumes 'terraform apply' is run from the 'terraform' directory (path.module is components/event-producer).
    # So path to event-producer is '../../event-producer'
    go_mod_hash     = sha1(file("${var.event_producer_source_path}/go.mod"))
    go_sum_hash     = sha1(file("${var.event_producer_source_path}/go.sum"))
    dockerfile_hash = sha1(file("${var.event_producer_source_path}/Dockerfile"))
    src_hash = sha1(join("", [
      for f in fileset("${var.event_producer_source_path}/", "**/*.go") :
      filemd5("${var.event_producer_source_path}/${f}")
    ]))
  }

  provisioner "local-exec" {
    # Build the Docker image.
    # The image is tagged with the name and tag provided by variables.
    command     = "docker build -t ${var.event_producer_image_name}:${var.event_producer_image_tag} ${var.event_producer_source_path}"
    # Working directory is implicitly the event-producer source path due to the command structure, or can be explicitly set.
    # working_dir = var.event_producer_source_path # Optional: Explicitly set if Dockerfile relies on CWD
  }
}

resource "kubernetes_namespace" "event_producer_ns" {
  count = var.create_namespace ? 1 : 0
  metadata {
    name = var.event_producer_namespace
  }
}

resource "kubernetes_config_map_v1" "event_producer_config" {
  metadata {
    name      = "event-producer-config"
    namespace = var.event_producer_namespace
  }

  data = {
    "KAFKA_BROKERS"  = var.kafka_brokers
    "KAFKA_TOPIC"    = var.kafka_topic
    "NUM_CITIZENS"   = tostring(var.num_citizens)
    "SEND_INTERVAL"  = var.send_interval
    "MONGO_URI"      = var.mongo_uri
    "DGRAPH_URL"     = var.dgraph_url
  }

  depends_on = [kubernetes_namespace.event_producer_ns]
}

resource "kubernetes_deployment_v1" "event_producer_deployment" {
  metadata {
    name      = "event-producer"
    namespace = var.event_producer_namespace
    labels = {
      app = "event-producer"
    }
  }

  spec {
    replicas = var.replicas
    selector {
      match_labels = {
        app = "event-producer"
      }
    }
    template {
      metadata {
        labels = {
          app = "event-producer"
        }
      }
      spec {
        container {
          name  = "event-producer"
          image = "${var.event_producer_image_name}:${var.event_producer_image_tag}"
          image_pull_policy = "IfNotPresent" # Important for locally built images / Kind/OrbStack

          env_from {
            config_map_ref {
              name = kubernetes_config_map_v1.event_producer_config.metadata[0].name
            }
          }

          resources {
            limits = {
              memory = "128Mi"
              cpu    = "250m"
            }
            requests = {
              memory = "64Mi"
              cpu    = "100m"
            }
          }
        }
      }
    }
  }

  depends_on = [
    null_resource.build_event_producer_image,
    kubernetes_config_map_v1.event_producer_config,
  ]
} 