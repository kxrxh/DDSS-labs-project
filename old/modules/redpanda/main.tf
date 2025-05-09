resource "helm_release" "redpanda" {
  name             = var.release_name
  namespace        = var.namespace
  create_namespace = var.create_namespace

  repository       = "https://charts.redpanda.com/"
  chart            = "redpanda"

  # Basic values for a dev setup. Customize as needed.
  # Refer to the Redpanda Helm chart documentation for all options:
  # https://github.com/redpanda-data/helm-charts/tree/main/charts/redpanda
  values = [
    yamlencode({
      # Use the specified storage class
      "storage" = {
        "persistentVolume" = {
          "storageClass" = var.storage_class_name
          "size" = "10Gi" # Adjust size as needed for dev
        }
      }
      # Set replica count
      "statefulset" = {
        "replicas" = var.replicas,
        "sideCars" = {
          "tuning" = {
            "enabled" = false
          }
        }
      }
      # External access configuration
      # This top-level 'external' block seems correct for NodePort/LB Service creation
      "external" = {
        "enabled" = true
        "type" = "NodePort" # Or LoadBalancer
      }
      "listeners" = {
        "kafka" = {
          # Correct structure for enabling external Kafka listener
          "external" = {
            "default" = { # A named external listener config
                "enabled" = true
                # Add specific port/advertised address config here if needed
            }
          },
          # Disable TLS for Kafka listeners
          "tls" = { "enabled" = false }
        }
        "admin" = {
          # Correct structure for enabling external admin listener
          "external" = {
             "default" = { # A named external listener config
                "enabled" = true
                # Add specific port/advertised address config here if needed
            }
          },
          # Disable TLS for Admin API
          "tls" = { "enabled" = false }
        }
        "schemaRegistry" = {
           "enabled" = false # Keep disabled for now
           # If enabling, external would follow the same object structure:
           # "external" = {
           #   "default" = { "enabled" = true }
           # }
        }
      }
      # Explicitly disable chart's cert-manager integration
      "tls" = {
        "enabled" = false
      }
      "tieredStorage" = {
        "enabled" = false
      }
      "resources" = {
        "cpu" = {
          "cores" = "1"
        }
        "memory" = {
          "container" = {
            "max" = "2Gi"
          }
        }
      }
      # Other settings like TLS can be configured here
      # "tls" = {
      #   "enabled" = true
      #   "kafka" = { ... }
      #   ...
      # }
    })
  ]

  # If Helm operations take time
  timeout = 600 # 10 minutes
}
