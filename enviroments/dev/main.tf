provider "kubernetes" {
  config_path = "~/.docker/orb/kube/config"  # OrbStack specific path
  config_context = "orbstack"
}

module "mongodb" {
  source = "../../modules/mongodb"

  namespace    = "mongodb-dev"
  replicas     = 1          # Fewer replicas in dev
  storage_size = "5Gi"      # Smaller storage in dev
  mongo_image = "mongo:5.0"
}