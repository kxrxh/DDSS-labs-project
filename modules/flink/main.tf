resource "helm_release" "flink_operator" {
  name             = var.release_name
  namespace        = var.namespace
  create_namespace = var.create_namespace

  repository       = try(var.flink_operator_helm_config["repository"], "https://downloads.apache.org/flink/flink-kubernetes-operator-${var.chart_version}/")
  chart            = "flink-kubernetes-operator"
  version          = var.chart_version

  values = [
    yamlencode({
      # Configure the operator to watch specific namespaces
      "watchNamespaces" = var.watch_namespaces
      # Disable webhook creation to avoid cert-manager dependency if not installed/running
      "webhook" = {
        "create" = false
      }
      # Add any other specific operator configurations here
      # Refer to the Flink Operator Helm chart documentation:
      # https://nightlies.apache.org/flink/flink-kubernetes-operator-docs-main/docs/operations/configuration/
    })
  ]

  timeout = 600 # 10 minutes
}

# Note: This module only deploys the Flink Operator.
# Actual Flink jobs (FlinkDeployment or FlinkSessionJob resources)
# need to be defined separately, typically as Kubernetes manifests
# or potentially using the kubernetes_manifest resource in Terraform
# (if the CRDs are installed by the operator first).
