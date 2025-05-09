resource "helm_release" "dgraph" {
  name       = var.dgraph_name
  repository = var.dgraph_repository
  chart      = var.dgraph_chart_name
  version    = var.dgraph_chart_version
  namespace  = var.dgraph_namespace
  create_namespace = true

  values = [
    <<-EOT
    # Dgraph Helm chart values
    # Refer to the Dgraph Helm chart documentation for available options:
    # https://github.com/dgraph-io/charts/tree/master/charts/dgraph

    # Example: Persist data (recommended for production)
    # alpha:
    #   persistence:
    #     enabled: true
    #     size: 10Gi # Adjust as needed
    # zero:
    #   persistence:
    #     enabled: true
    #     size: 1Gi  # Adjust as needed

    # For a minimal setup, you might not need to override many values.
    # Check the default values.yaml in the Dgraph chart.

    alpha:
      replicaCount: 2
    zero:
      replicaCount: 2
    EOT
  ]

  # It's common to need to set specific configurations for Dgraph Alpha and Zero nodes.
  # For example, to run more Alpha nodes for higher read/write throughput or more Zero nodes for HA.
  # set {
  #   name  = "alpha.replicas"
  #   value = "1" # Default is 1, adjust if needed
  # }
  # set {
  #   name  = "zero.replicas"
  #   value = "1" # Default is 1, for HA you'd typically use 3
  # }
} 