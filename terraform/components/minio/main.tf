resource "kubernetes_namespace" "minio_ns" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_secret" "minio_credentials" {
  metadata {
    name      = "minio-credentials"
    namespace = var.namespace
  }

  data = {
    root-user     = var.minio_root_user
    root-password = var.minio_root_password
  }

  type = "Opaque"
}

resource "kubernetes_service" "minio_service" {
  metadata {
    name      = "minio-service"
    namespace = var.namespace
    labels = {
      app = "minio"
    }
  }
  spec {
    type = "ClusterIP"
    port {
      name        = "api"
      port        = 9000
      target_port = 9000
    }
    port {
      name        = "console"
      port        = 9001
      target_port = 9001
    }
    selector = {
      app = "minio"
    }
  }
}

resource "kubernetes_deployment" "minio" {
  metadata {
    name      = "minio"
    namespace = var.namespace
    labels = {
      app = "minio"
    }
  }

  spec {
    replicas = var.replicas

    selector {
      match_labels = {
        app = "minio"
      }
    }

    template {
      metadata {
        labels = {
          app = "minio"
        }
      }

      spec {
        container {
          name  = "minio"
          image = var.minio_image

          command = ["minio"]
          args    = ["server", "/data", "--console-address", ":9001"]

          port {
            container_port = 9000
          }
          port {
            container_port = 9001
          }

          env {
            name = "MINIO_ROOT_USER"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.minio_credentials.metadata[0].name
                key  = "root-user"
              }
            }
          }

          env {
            name = "MINIO_ROOT_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.minio_credentials.metadata[0].name
                key  = "root-password"
              }
            }
          }

          volume_mount {
            name       = "minio-data"
            mount_path = "/data"
          }

          liveness_probe {
            http_get {
              path = "/minio/health/live"
              port = 9000
            }
            initial_delay_seconds = 30
            period_seconds        = 30
          }

          readiness_probe {
            http_get {
              path = "/minio/health/ready"
              port = 9000
            }
            initial_delay_seconds = 15
            period_seconds        = 15
          }

          resources {
            requests = {
              memory = var.memory_request
              cpu    = var.cpu_request
            }
            limits = {
              memory = var.memory_limit
              cpu    = var.cpu_limit
            }
          }
        }

        volume {
          name = "minio-data"
          
          # Use persistent storage if enabled, otherwise use emptyDir for development
          dynamic "persistent_volume_claim" {
            for_each = var.use_persistent_storage ? [1] : []
            content {
              claim_name = kubernetes_persistent_volume_claim.minio_pvc[0].metadata[0].name
            }
          }
          
          dynamic "empty_dir" {
            for_each = var.use_persistent_storage ? [] : [1]
            content {
              size_limit = var.storage_size
            }
          }
        }
      }
    }
  }
}

resource "kubernetes_persistent_volume_claim" "minio_pvc" {
  count = var.use_persistent_storage ? 1 : 0
  
  metadata {
    name      = "minio-data-pvc"
    namespace = var.namespace
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = var.storage_size
      }
    }
    # Only set storage class if provided, otherwise use default
    storage_class_name = var.storage_class != "" ? var.storage_class : null
  }
}

# Note: Buckets are now created automatically by the MinIO sink when needed
# No initialization job required! 