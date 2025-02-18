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
  replicas     = 1          # Fewer replicas in dev
  storage_size = "5Gi"      # Smaller storage in dev
  mongo_image = "mongo:latest"
} 