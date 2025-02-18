terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"  # ABSOLUTE PATH -  Change this!
}

module "mongodb" {
  source = "../../modules/mongodb"

  namespace    = "mongodb-dev"
  replicas     = 2          # Fewer replicas in dev
  storage_size = "5Gi"      # Smaller storage in dev
  mongo_image = "mongo:latest"
}

module "dgraph" {
  source = "../../modules/dgraph"

  namespace          = "dgraph-dev"
  zero_replicas      = 1
  alpha_replicas     = 1
  zero_storage_size  = "2Gi"  # Smaller storage in dev
  alpha_storage_size = "2Gi"  # Smaller storage in dev
} 