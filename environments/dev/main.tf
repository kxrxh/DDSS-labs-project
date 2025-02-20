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
  replicas     = 3          # Minimum for proper replica set elections
  storage_size = "5Gi"
  mongo_image = "mongo:latest"
}

module "dgraph" {
  source = "../../modules/dgraph"

  namespace          = "dgraph-dev"
  zero_replicas      = 3     # Odd number for raft consensus
  alpha_replicas     = 2     # Even number can accept rolling updates
  zero_storage_size  = "2Gi" 
  alpha_storage_size = "2Gi" 
} 

module "tikv" {
  source       = "../../modules/tikv"
  namespace    = "tikv-dev"
  pd_replicas  = 1
  tikv_replicas = 1
  pd_image     = "pingcap/pd:latest"
  tikv_image   = "pingcap/tikv:latest"
}