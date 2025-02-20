terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

module "mongodb" {
  source = "../../modules/mongodb"

  namespace    = "mongodb-dev"
  replicas     = 5          # Updated
  storage_size = "5Gi"      # Increase based on data size
  mongo_image  = "mongo:latest"
}

module "dgraph" {
  source = "../../modules/dgraph"

  namespace          = "dgraph-dev"
  zero_replicas      = 5     # Updated
  alpha_replicas     = 6     # Updated
  zero_storage_size  = "2Gi" 
  alpha_storage_size = "2Gi" 
}

module "tikv" {
  source        = "../../modules/tikv"
  namespace     = "tikv-dev"
  pd_replicas   = 1     # Updated
  tikv_replicas = 1     # Updated
  pd_image      = "pingcap/pd:latest"
  tikv_image    = "pingcap/tikv:latest"
}