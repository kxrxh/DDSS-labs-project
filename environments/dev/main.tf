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
    # Add other providers if needed (e.g., aws, google, azure)
  }
}

# Configure Kubernetes provider (adjust context if needed)
provider "kubernetes" {
  # Assumes kubeconfig is set up correctly in the environment
  config_path    = "~/.kube/config"
  # config_context = "your-dev-cluster-context" # Uncomment and set if you use a non-default context
}

# Configure Helm provider to use the Kubernetes provider context
provider "helm" {
  kubernetes {
    # Assumes kubeconfig is set up correctly in the environment
    config_path    = "~/.kube/config"
    # config_context = "your-dev-cluster-context" # Uncomment and set if you use a non-default context
  }
}

resource "kubernetes_namespace" "database" {
  metadata {
    # Use a consistent name for the database namespace
    name = "dev-db"
  }
}

module "mongodb" {
  source = "../../modules/mongodb"
  namespace = kubernetes_namespace.database.metadata[0].name
  # TODO: Ensure mongodb module does not create the namespace itself
}

module "dgraph" {
  source = "../../modules/dgraph"
  namespace = kubernetes_namespace.database.metadata[0].name
  # TODO: Ensure dgraph module does not create the namespace itself
}

module "redpanda" {
  source = "../../modules/redpanda"

  namespace = "dev-streaming" # Dedicated namespace for streaming components
  release_name = "dev-redpanda"
  replicas = 1 # Keep it small for dev
  # Use the available StorageClass found in the cluster
  storage_class_name = "local-path"
  create_namespace = true
}

module "flink_operator" {
  source = "../../modules/flink"

  namespace = "dev-flink-operator" # Separate namespace for the operator itself
  release_name = "dev-flink-operator"
  watch_namespaces = ["default", "dev-flink-jobs"] # Operator will watch these namespaces for Flink jobs
  create_namespace = true
}

# TODO: Define FlinkDeployment/FlinkSessionJob resources in the 'dev-flink-jobs' namespace
# This could be done via separate Kubernetes manifests (kubectl apply)
# or using kubernetes_manifest resources here after the operator is ready.

# Example of creating the namespace for jobs (optional, chart can create too)
# resource "kubernetes_namespace" "flink_jobs" {
#   metadata {
#     name = "dev-flink-jobs"
#   }
# }