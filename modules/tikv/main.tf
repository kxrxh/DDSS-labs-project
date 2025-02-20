# Namespace
resource "kubernetes_namespace" "tikv" {
  metadata {
    name = var.namespace
  }
}

# Placement Driver (PD) Service
resource "kubernetes_service" "pd_service" {
  metadata {
    name      = "tikv-pd"
    namespace = var.namespace
  }
  spec {
    selector = {
      app = "tikv-pd"
    }
    port {
      name        = "pd-client"
      port        = 2379
      target_port = 2379
    }
    cluster_ip = "None" # Headless service for PD
  }
}

# TiKV Service (unchanged)
resource "kubernetes_service" "tikv_service" {
  metadata {
    name      = "tikv"
    namespace = var.namespace
  }
  spec {
    selector = {
      app = "tikv"
    }
    port {
      name        = "tikv-client"
      port        = 20160
      target_port = 20160
    }
    cluster_ip = "None" # Headless service for TiKV
  }
}

# Placement Driver (PD) StatefulSet
resource "kubernetes_stateful_set" "pd" {
  metadata {
    name      = "tikv-pd"
    namespace = var.namespace
  }
  spec {
    service_name = "tikv-pd"
    replicas     = var.pd_replicas
    selector {
      match_labels = {
        app = "tikv-pd"
      }
    }
    template {
      metadata {
        labels = {
          app = "tikv-pd"
        }
      }
      spec {
        container {
          name  = "pd"
          image = var.pd_image
          args  = [
            "--name=$(POD_NAME)",
            "--client-urls=http://0.0.0.0:2379",  # Listen on all interfaces for clients
            "--peer-urls=http://0.0.0.0:2380",    # Listen on all interfaces for peers
            "--advertise-client-urls=http://$(POD_NAME).${kubernetes_service.pd_service.metadata[0].name}.${var.namespace}.svc.cluster.local:2379",
            "--advertise-peer-urls=http://$(POD_NAME).${kubernetes_service.pd_service.metadata[0].name}.${var.namespace}.svc.cluster.local:2380",
            "--initial-cluster=${join(",", [for i in range(var.pd_replicas) : format("tikv-pd-%d=http://tikv-pd-%d.%s.%s.svc.cluster.local:2380", i, i, kubernetes_service.pd_service.metadata[0].name, var.namespace)])}",
            "--data-dir=/data/pd"
          ]
          env {
            name = "POD_NAME"
            value_from {
              field_ref {
                field_path = "metadata.name"
              }
            }
          }
          port {
            container_port = 2379
            name           = "pd-client"
          }
          port {
            container_port = 2380
            name           = "pd-peer"
          }
          volume_mount {
            name       = "pd-storage"
            mount_path = "/data/pd"
          }
          readiness_probe {
            http_get {
              path = "/health"
              port = 2379
            }
            initial_delay_seconds = 5
          }
        }
      }
    }
    volume_claim_template {
      metadata {
        name = "pd-storage"
      }
      spec {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = var.pd_storage_size
          }
        }
      }
    }
  }
}

# TiKV StatefulSet (unchanged for now)
resource "kubernetes_stateful_set" "tikv" {
  metadata {
    name      = "tikv"
    namespace = var.namespace
  }
  spec {
    service_name = "tikv"
    replicas     = var.tikv_replicas
    selector {
      match_labels = {
        app = "tikv"
      }
    }
    template {
      metadata {
        labels = {
          app = "tikv"
        }
      }
      spec {
        container {
          name  = "tikv"
          image = var.tikv_image
          args  = [
            "--pd=${kubernetes_service.pd_service.metadata[0].name}.${var.namespace}.svc.cluster.local:2379",
            "--addr=0.0.0.0:20160",
            "--data-dir=/data/tikv",
            "--advertise-addr=$(POD_NAME).${kubernetes_service.tikv_service.metadata[0].name}.${var.namespace}.svc.cluster.local:20160"
          ]
          env {
            name = "POD_NAME"
            value_from {
              field_ref {
                field_path = "metadata.name"
              }
            }
          }
          port {
            container_port = 20160
          }
          volume_mount {
            name       = "tikv-storage"
            mount_path = "/data/tikv"
          }
        }
      }
    }
    volume_claim_template {
      metadata {
        name = "tikv-storage"
      }
      spec {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = var.tikv_storage_size
          }
        }
      }
    }
  }
}