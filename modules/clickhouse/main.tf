# This resource now installs the main 'clickhouse' chart,
# which includes the operator as a dependency by default.
resource "helm_release" "clickhouse_operator" {
  # Renamed release to 'dev-clickhouse' as it manages the main installation now.
  name             = var.operator_release_name 
  # Namespace for BOTH the ClickHouse cluster AND the operator.
  namespace        = var.installation_namespace
  create_namespace = var.operator_create_namespace # Allow chart to create the installation namespace

  repository       = "https://helm.altinity.com"
  # Install the main 'clickhouse' chart
  chart            = "clickhouse"
  # Use latest version by default (variable default is null)
  version          = var.operator_chart_version

  values = [
    # Configure the ClickHouseInstallation directly within the chart's values
    yamlencode({

      # Explicitly name the ClickHouseInstallation resource 
      # (defaults to release name if omitted, but good to be explicit)
      # chi_name = var.operator_release_name 

      # Tell the chart to create the ClickHouseInstallation resource
      clickhouse = {
        create = true 
        # Define the spec for the ClickHouseInstallation resource
        spec = {
          # Use the configuration section previously in the manifest
          configuration = {
            # Restore users block
            users = {
              "kxrxh/password" = "mycoolpassword"
              # "default/password" = "SomeStrongPassword" 
            }
            # Add other settings like zookeeper, profiles if needed
          }
          # Use the defaults section previously in the manifest
          defaults = {
            templates = {
              podTemplate = "dev-clickhouse-pod"
              dataVolumeClaimTemplate = "dev-clickhouse-data"
              serviceTemplate = "dev-clickhouse-service"
            }
          }
          # Use the templates section previously in the manifest
          templates = {
            podTemplates = [
              {
                name = "dev-clickhouse-pod"
                spec = {
                  containers = [
                    {
                      name = "clickhouse"
                      # Add resources here if needed
                    }
                  ]
                }
              }
            ]
            volumeClaimTemplates = [
              {
                name = "dev-clickhouse-data"
                spec = {
                  accessModes = ["ReadWriteOnce"]
                  resources = { requests = { storage = "1Gi" } }
                  storageClassName = "local-path"
                }
              }
            ]
            serviceTemplates = [
              {
                name = "dev-clickhouse-service"
                spec = {
                  type = "ClusterIP"
                  ports = [
                    { name = "http", port = 8123 },
                    { name = "tcp", port = 9000 }
                  ]
                }
              }
            ]
          }
        }
      }
      # You can add operator-specific configurations here if needed, 
      # outside the 'clickhouse.spec' block.
    })
  ]

  timeout = 600
}

# Removed the separate kubernetes_manifest resource for ClickHouseInstallation.
# The 'helm_release' resource above now manages the CHI creation via its values. 