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

# TiKV Service
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

# Placement Driver (PD) Deployment
resource "kubernetes_deployment" "pd" {
  metadata {
    name      = "tikv-pd"
    namespace = var.namespace
  }
  spec {
    replicas = var.pd_replicas
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
            "--name=pd",
            "--advertise-client-urls=http://tikv-pd.${var.namespace}.svc.cluster.local:2379",
            "--advertise-peer-urls=http://tikv-pd.${var.namespace}.svc.cluster.local:2380",
            "--initial-cluster=pd=http://tikv-pd.${var.namespace}.svc.cluster.local:2380",
            "--data-dir=/data/pd"
          ]
          port {
            container_port = 2379
          }
          volume_mount {
            name       = "pd-storage"
            mount_path = "/data/pd"
          }
        }
        volume {
          name = "pd-storage"
        }
      }
    }
  }
}

# PD Persistent Volume Claim
resource "kubernetes_persistent_volume_claim" "pd_pvc" {
  metadata {
    name      = "tikv-pd-pvc"
    namespace = var.namespace
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

# TiKV Deployment
resource "kubernetes_deployment" "tikv" {
  metadata {
    name      = "tikv"
    namespace = var.namespace
  }
  spec {
    replicas = var.tikv_replicas
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
            "--pd=tikv-pd.${var.namespace}.svc.cluster.local:2379",
            "--addr=0.0.0.0:20160",
            "--data-dir=/data/tikv"
          ]
          port {
            container_port = 20160
          }
          volume_mount {
            name       = "tikv-storage"
            mount_path = "/data/tikv"
          }
        }
        volume {
          name = "tikv-storage"
        }
      }
    }
  }
}

# TiKV Persistent Volume Claim
resource "kubernetes_persistent_volume_claim" "tikv_pvc" {
  metadata {
    name      = "tikv-pvc"
    namespace = var.namespace
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