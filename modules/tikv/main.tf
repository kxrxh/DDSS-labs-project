# modules/tikv/main.tf


# Create the namespace
resource "kubernetes_namespace" "tikv_namespace" {
  metadata {
    name = var.namespace
  }
}

# PD StatefulSet
resource "kubernetes_stateful_set" "pd" {
  depends_on = [kubernetes_namespace.tikv_namespace]

  metadata {
    name      = "pd"
    namespace = var.namespace
  }

  spec {
    replicas = var.pd_replicas
    service_name = "pd-service"

    selector {
      match_labels = {
        app = "pd"
      }
    }

    template {
      metadata {
        labels = {
          app = "pd"
        }
      }

      spec {
        container {
          image = var.pd_image
          name  = "pd"

          args = [
            "--name=$(POD_NAME)",
            "--client-urls=http://0.0.0.0:2379",
            "--peer-urls=http://0.0.0.0:2380",
            "--advertise-client-urls=http://$(POD_NAME).pd-service.${var.namespace}.svc.cluster.local:2379",
            "--advertise-peer-urls=http://$(POD_NAME).pd-service.${var.namespace}.svc.cluster.local:2380",
            "--initial-cluster=${join(",", [for i in range(var.pd_replicas) : "pd-${i}=http://pd-${i}.pd-service.${var.namespace}.svc.cluster.local:2380"])}"
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
            container_port = 2379  # Client port
          }
          port {
            container_port = 2380  # Peer port
          }
        }
      }
    }
  }
}

# PD Headless Service
resource "kubernetes_service" "pd_service" {
  depends_on = [kubernetes_namespace.tikv_namespace]

  metadata {
    name      = "pd-service"
    namespace = var.namespace
  }

  spec {
    selector = {
      app = "pd"
    }

    port {
      port        = 2379
      target_port = 2379
      name        = "client"
    }

    port {
      port        = 2380
      target_port = 2380
      name        = "peer"
    }

    cluster_ip = "None"  # Headless service
  }
}

# TiKV StatefulSet (replacing Deployment)
resource "kubernetes_stateful_set" "tikv" {
  depends_on = [kubernetes_namespace.tikv_namespace]

  metadata {
    name      = "tikv"
    namespace = var.namespace
  }

  spec {
    replicas = var.tikv_replicas
    service_name = "tikv-service"

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
          image = var.tikv_image
          name  = "tikv"

          args = [
            "--addr=0.0.0.0:20160",
            "--advertise-addr=$(POD_NAME).tikv-service.${var.namespace}.svc.cluster.local:20160",
            "--pd-endpoints=pd-service.${var.namespace}.svc.cluster.local:2379"
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
            container_port = 20160  # TiKV client port
          }
        }
      }
    }
  }
}

# TiKV Headless Service
resource "kubernetes_service" "tikv_service" {
  depends_on = [kubernetes_namespace.tikv_namespace]

  metadata {
    name      = "tikv-service"
    namespace = var.namespace
  }

  spec {
    selector = {
      app = "tikv"
    }

    port {
      port        = 20160
      target_port = 20160
      name        = "client"
    }

    cluster_ip = "None"  # Headless service for stable DNS
  }
}