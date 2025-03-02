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
  source       = "../../modules/mongodb"
  namespace    = "mongodb-dev"
  replicas     = 1          # Reduced to standard HA
  storage_size = "10Gi"    # Increased for realistic data
  mongo_image  = "mongo:latest"
}

module "dgraph" {
  source             = "../../modules/dgraph"
  namespace          = "dgraph-dev"
  zero_replicas      = 1    # Reduced to standard HA
  alpha_replicas     = 1    # Start with 3, scale if needed
  zero_storage_size  = "2Gi" 
  alpha_storage_size = "2Gi" # Increased for data
}

module "tikv" {
  source        = "../../modules/tikv"
  namespace     = "tikv-dev"
  pd_replicas   = 1     # Reduced to standard HA
  tikv_replicas = 1     # Start with 3, scale if needed
  pd_image      = "pingcap/pd:latest"
  tikv_image    = "pingcap/tikv:latest"
}