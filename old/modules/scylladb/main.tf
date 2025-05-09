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

resource "kubernetes_namespace" "scylladb" {
  count = var.create_namespace ? 1 : 0
  metadata {
    name = var.namespace
  }
}

resource "helm_release" "scylladb" {
  name       = var.release_name
  repository = "https://scylla-operator-charts.storage.googleapis.com/stable"
  chart      = "scylla"
  version    = var.chart_version # Specify a version for consistency
  namespace  = var.namespace

  # Values for a lean, local setup
  # Refer to ScyllaDB Helm chart documentation for all options:
  # https://github.com/scylladb/scylla-operator/tree/master/deploy/charts/scylla
  set {
    name  = "developerMode"
    value = "true"
  }

  set {
    name  = "cpuset" # Disable cpuset for local dev, not typically needed/problematic in kind/minikube
    value = "false"
  }

  set {
    name  = "scylla.agent.enabled" # Disable Scylla agent sidecar for local dev
    value = false
  }

  set {
    name = "scylla.scyllaManager.enabled" # Attempt to disable Scylla Manager integration
    value = false
  }
  
  set {
    name = "scylla.storageClass" # Use the provided storage class
    value = var.persistence_storage_class
  }

  set {
    name = "scylla.persistentVolume.size"
    value = var.persistence_size
  }

  set {
    name = "scylla.datacenter.racks[0].members" # Single node for local dev
    value = "1" 
  }

  # Ensure namespace exists before trying to install chart
  depends_on = [kubernetes_namespace.scylladb]
} 