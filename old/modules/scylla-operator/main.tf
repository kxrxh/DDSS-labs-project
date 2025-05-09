terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.27"
    }
  }
}

resource "kubernetes_namespace" "scylla_operator" {
  count = var.create_namespace ? 1 : 0
  metadata {
    name = var.namespace
  }
}

resource "helm_release" "scylla_operator" {
  name       = var.release_name
  repository = "https://scylla-operator-charts.storage.googleapis.com/stable"
  chart      = "scylla-operator"
  version    = var.chart_version
  namespace  = var.namespace

  # Ensure namespace exists before trying to install chart
  # and respect explicit dependencies passed to the module
  depends_on = [kubernetes_namespace.scylla_operator]
} 