resource "kubernetes_namespace" "dgraph" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_stateful_set" "dgraph_zero" {
  metadata {
    name      = "dgraph-zero"
    namespace = var.namespace
  }
  spec {
    service_name = "dgraph-zero"
    replicas     = var.zero_replicas
    selector {
      match_labels = {
        app = "dgraph-zero"
      }
    }
    template {
      metadata {
        labels = {
          app = "dgraph-zero"
        }
      }
      spec {
        container {
          name  = "dgraph-zero"
          image = var.dgraph_zero_image

          port {
            container_port = 6080
            name           = "grpc"
          }
          port {
            container_port = 8080
            name           = "http"
          }
          volume_mount {
            mount_path = "/dgraph"
            name       = "dgraph-zero-data"
          }
          command = ["dgraph", "zero", "--my=dgraph-zero:5080"]
        }
      }
    }
    volume_claim_template {
      metadata {
        name = "dgraph-zero-data"
      }
      spec {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = var.zero_storage_size
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "dgraph_zero" {
  metadata {
    name      = "dgraph-zero"
    namespace = var.namespace
    labels = {
      app = "dgraph-zero"
    }
  }
  spec {
    port {
      port        = 6080
      target_port = "grpc"
      name        = "grpc"
    }
    port {
      port        = 8080
      target_port = "http"
      name        = "http"
    }
    cluster_ip = "None"
    selector = {
      app = "dgraph-zero"
    }
  }
}

resource "kubernetes_stateful_set" "dgraph_alpha" {
  metadata {
    name      = "dgraph-alpha"
    namespace = var.namespace
  }
  spec {
    service_name = "dgraph-alpha"
    replicas     = var.alpha_replicas
    selector {
      match_labels = {
        app = "dgraph-alpha"
      }
    }
    template {
      metadata {
        labels = {
          app = "dgraph-alpha"
        }
      }
      spec {
        container {
          name  = "dgraph-alpha"
          image = var.dgraph_alpha_image

          port {
            container_port = 7080
            name           = "grpc-internal"
          }
          port {
            container_port = 9080
            name           = "grpc"
          }
          port {
            container_port = 8080
            name           = "http"
          }
          volume_mount {
            mount_path = "/dgraph"
            name       = "dgraph-alpha-data"
          }
          command = [
            "dgraph",
            "alpha",
            "--my=dgraph-alpha:7080",
            "--zero=dgraph-zero:5080",
          ]
        }
      }
    }
    volume_claim_template {
      metadata {
        name = "dgraph-alpha-data"
      }
      spec {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = var.alpha_storage_size
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "dgraph_alpha" {
  metadata {
    name      = "dgraph-alpha"
    namespace = var.namespace
    labels = {
      app = "dgraph-alpha"
    }
  }
  spec {
    port {
      port        = 9080
      target_port = "grpc"
      name        = "grpc"
    }
    port {
      port        = 8080
      target_port = "http"
      name        = "http"
    }
    selector = {
      app = "dgraph-alpha"
    }
  }
}

# Deploy Ratel Dashboard
resource "kubernetes_deployment" "dgraph_ratel" {
  metadata {
    name      = "dgraph-ratel"
    namespace = var.namespace
  }
  spec {
    replicas = var.ratel_replicas
    selector {
      match_labels = {
        app = "dgraph-ratel"
      }
    }
    template {
      metadata {
        labels = {
          app = "dgraph-ratel"
        }
      }
      spec {
        container {
          name  = "dgraph-ratel"
          image = var.dgraph_ratel_image
          
          port {
            container_port = 8000
            name           = "http"
          }
          
          resources {
            limits = {
              cpu    = var.ratel_cpu_limit
              memory = var.ratel_memory_limit
            }
            requests = {
              cpu    = var.ratel_cpu_request
              memory = var.ratel_memory_request
            }
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "dgraph_ratel" {
  metadata {
    name      = "dgraph-ratel"
    namespace = var.namespace
    labels = {
      app = "dgraph-ratel"
    }
  }
  spec {
    port {
      port        = 8000
      target_port = "http"
      name        = "http"
    }
    selector = {
      app = "dgraph-ratel"
    }
    type = var.ratel_service_type
  }
}
