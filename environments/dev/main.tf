terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "kubernetes" {
  config_path = "/Users/kxrxh/.kube/config"  # ABSOLUTE PATH -  Change this!
}

module "mongodb" {
  source = "../../modules/mongodb"

  namespace    = "mongodb-dev-new"
  replicas     = 1          # Fewer replicas in dev
  storage_size = "5Gi"      # Smaller storage in dev
  mongo_image = "mongo:5.0"
} 