terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

resource "kubernetes_namespace" "mongodb" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_service" "mongodb_headless" {
  metadata {
    name      = "mongodb-service"
    namespace = kubernetes_namespace.mongodb.metadata[0].name
    labels = {
      app = "mongodb"
    }
  }
  spec {
    cluster_ip = "None"
    port {
      port        = 27017
      target_port = 27017
    }
    selector = {
      app = "mongodb"
    }
  }
}

resource "kubernetes_stateful_set" "mongodb" {
  metadata {
    name      = "mongodb"
    namespace = kubernetes_namespace.mongodb.metadata[0].name
  }

  spec {
    service_name = kubernetes_service.mongodb_headless.metadata[0].name
    replicas     = var.replicas

    selector {
      match_labels = {
        app = "mongodb"
      }
    }

    template {
      metadata {
        labels = {
          app = "mongodb"
        }
      }

      spec {
        container {
          name  = "mongodb"
          image = var.mongo_image

          port {
            container_port = 27017
          }

          volume_mount {
            name       = "mongodb-data"
            mount_path = "/data/db"
          }
        }
      }
    }

    volume_claim_template {
      metadata {
        name = "mongodb-data"
      }

      spec {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = var.storage_size
          }
        }
      }
    }
  }
}

# ... (rest of the MongoDB resources from the previous example) ...
# Parameterize everything using variables (see variables.tf below) 