
provider "kubernetes" {
  config_path = "~/.kube/config"
}

provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }
}

resource "helm_release" "influxdb" {
  name             = var.release_name
  namespace        = var.namespace
  create_namespace = var.create_namespace

  repository       = "https://helm.influxdata.com/"
  chart            = "influxdb2"

  values = [
    yamlencode({
      persistence = { 
        enabled = var.persistence_enabled
        size    = var.persistence_size
        # Set storageClassName only if specified, otherwise let Helm/K8s use default
        storageClassName = var.persistence_storage_class != null ? var.persistence_storage_class : ""
      }
      # Admin user configuration
      adminUser = {
        username = var.admin_username
        password = var.admin_password
      }
      # Expose service configuration
      service = {
        type = "ClusterIP"
        port = 8086
      }
      # Resource limits
      resources = {
        limits = {
          cpu    = "1000m"
          memory = "2Gi"
        }
        requests = {
          cpu    = "100m"
          memory = "512Mi"
        }
      }
    })
  ]

  timeout = 600 # 10 minutes
}



# Create a cleanup job for development purposes
resource "kubernetes_job" "influxdb_cleanup" {
  count = var.enable_cleanup_job ? 1 : 0

  metadata {
    name      = "${var.release_name}-cleanup"
    namespace = var.namespace
  }

  spec {
    template {
      metadata {
        labels = {
          app = "${var.release_name}-cleanup"
        }
      }

      spec {
        restart_policy = "OnFailure"

        container {
          name  = "influxdb-cleanup"
          image = "alpine:latest"

          command = ["/bin/sh"]
          args = ["-c", <<-EOF
            apk add --no-cache jq
            
            echo "Cleaning up old data..."
            
            # Get admin token from secret
            ADMIN_TOKEN="${var.admin_token}"
            
            # Delete and recreate buckets to clean data
            echo "Recreating buckets..."
            
            # List and delete existing buckets (except system buckets)
            BUCKETS=$(curl -s -X GET \
              "http://${var.release_name}-influxdb2.${var.namespace}.svc.cluster.local:8086/api/v2/buckets" \
              -H "Authorization: Token $ADMIN_TOKEN" \
              | jq -r '.buckets[] | select(.name != "_monitoring" and .name != "_tasks") | .id')
            
            for bucket_id in $BUCKETS; do
              echo "Deleting bucket: $bucket_id"
              curl -s -X DELETE \
                "http://${var.release_name}-influxdb2.${var.namespace}.svc.cluster.local:8086/api/v2/buckets/$bucket_id" \
                -H "Authorization: Token $ADMIN_TOKEN"
            done
            
            echo "Cleanup completed!"
          EOF
          ]
        }
      }
    }

    backoff_limit = 2
  }

  depends_on = [helm_release.influxdb]
} 