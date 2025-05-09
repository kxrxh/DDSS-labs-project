terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.36.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.17.0"
    }
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "orbstack"
}

provider "helm" {
  kubernetes {
    config_path    = "~/.kube/config"
    config_context = "orbstack"
  }
}

module "mongodb" {
  source = "./components/mongodb"

  # You can override default MongoDB settings here if needed, for example:
  # mongodb_namespace    = "mongo-apps"
  # mongodb_auth_enabled = true
  # mongodb_chart_version = "15.5.0" # Pin to a specific chart version
}

module "redpanda" {
  source = "./components/redpanda"

  # You can override default Redpanda settings here if needed, for example:
  # redpanda_namespace       = "redpanda-cluster"
  # redpanda_chart_version   = "5.6.5" # Pin to a specific chart version, ensure this matches your needs
  # redpanda_console_enabled = true

  # Example of how you might set values based on the Ansible variables provided earlier:
  # Note: These would typically be set within the module's values.yaml or via 'set' blocks if not directly mapped to variables.
  # This is just for conceptual illustration; actual Helm chart values will differ in structure.

  # redpanda_tiered_storage_bucket_name = "my-redpanda-backup-bucket" # If you want to enable tiered storage
}

module "dgraph" {
  source = "./components/dgraph"

  # You can override default Dgraph settings here if needed, for example:
  # dgraph_namespace       = "dgraph-system"
  # dgraph_chart_version   = "24.1.1" # Pin to a specific chart version
  # dgraph_alpha_replicas  = 3
  # dgraph_zero_replicas   = 3 # For High Availability
}

module "scylladb" {
  source = "./components/scylladb"

  # You can override default ScyllaDB settings here if needed, for example:
  # scylla_namespace = "scylla-prod"
  # scylla_operator_chart_version = "1.12.1" # Pin to a specific chart version
  # scylla_cluster_members = 3
  # scylla_cluster_cpus = "2"
  # scylla_cluster_memory = "4Gi"
} 

