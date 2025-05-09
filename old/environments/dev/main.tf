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

resource "kubernetes_namespace" "analytics" {
  metadata {
    name = "dev-analytics"
  }
}

# Create a dummy agent config secret to satisfy the ScyllaDB Operator
resource "kubernetes_secret" "scylla_agent_config_dummy" {
  metadata {
    name      = "scylla-agent-config"
    namespace = kubernetes_namespace.database.metadata[0].name
  }
  data = {
    "SCYLLA_AGENT_AUTH_TOKEN" = "dummytoken"
  }
  type = "Opaque"

  # Ensure the namespace exists before creating the secret
  depends_on = [kubernetes_namespace.database]
}

module "influxdb" {
  source = "../../modules/influxdb"

  namespace        = kubernetes_namespace.database.metadata[0].name
  release_name     = "dev-influxdb"
  # Let the module create the namespace if needed (redundant but safe)
  # create_namespace = true 
  persistence_storage_class = "local-path" # Use the same SC as redpanda
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

# Install cert-manager before ScyllaDB Operator
module "cert_manager" {
  source = "../../modules/cert-manager"
}

# Install ScyllaDB Operator, depends on cert-manager
module "scylla_operator" {
  source = "../../modules/scylla-operator"
  # Default namespace is scylla-operator, default release name is scylla-operator
  # Chart version defaults to a recent stable one in the module.
  depends_on = [module.cert_manager]
}

module "scylladb" {
  source = "../../modules/scylladb"

  namespace                 = kubernetes_namespace.database.metadata[0].name
  release_name              = "dev-scylladb"
  persistence_storage_class = "local-path" # Ensure this matches your available SC
  create_namespace = false

  # ScyllaDB cluster depends on the ScyllaDB Operator being ready
  # and the dummy agent config secret existing.
  depends_on = [
    module.scylla_operator,
    kubernetes_namespace.database,
    kubernetes_secret.scylla_agent_config_dummy
  ]
}

module "redpanda" {
  source = "../../modules/redpanda"

  namespace = "dev-streaming" # Dedicated namespace for streaming components
  release_name = "dev-redpanda"
  replicas = 1 # Keep it small for dev
  # Use the available StorageClass found in the cluster
  storage_class_name = "local-path"
  create_namespace = true

  # AND the dependent service modules completing their apply.
  depends_on = [
    module.flink_operator,
    null_resource.build_and_copy_flink_job,
    kubernetes_namespace.flink_jobs,
    module.scylladb, # Added dependency
    module.mongodb,
    module.dgraph,
    module.influxdb
  ]
}

module "flink_operator" {
  source = "../../modules/flink"

  namespace = "dev-flink-operator" # Separate namespace for the operator itself
  release_name = "dev-flink-operator"
  watch_namespaces = ["default", "dev-flink-jobs"] # Operator will watch these namespaces for Flink jobs
  create_namespace = true
}

# Ensure the Flink job JAR is built and copied to the expected host path
resource "null_resource" "build_and_copy_flink_job" {
  # Trigger this resource when files in the Flink job source directory change
  triggers = {
    # Hash of pom.xml and all Java source files
    source_code_hash = sha1(join("", [
      filemd5("${path.root}/../../social-rating-flink-job/pom.xml"),
      join("", [for f in fileset("${path.root}/../../social-rating-flink-job/src", "**/*.java") : filemd5("${path.root}/../../social-rating-flink-job/src/${f}")])
    ]))
  }

  provisioner "local-exec" {
    # Build command
    command     = "mvn clean package"
    working_dir = "${path.root}/../../social-rating-flink-job"
  }

  provisioner "local-exec" {
    # Copy command: create dir and copy JAR to the host path mounted by pods
    # Paths are relative to the environments/dev directory where terraform runs
    command = <<-EOT
      mkdir -p /tmp/flink-jars && \
      cp ../../social-rating-flink-job/target/social-rating-flink-job-1.0-SNAPSHOT.jar /tmp/flink-jars/social-rating-flink-job.jar
    EOT
  }
}

# Define the namespace for Flink jobs
resource "kubernetes_namespace" "flink_jobs" {
  metadata {
    name = "dev-flink-jobs" # Must match the namespace in social-rating-job.yaml and flink_operator watchNamespaces
  }
}

# Deploy the Flink job using the Kubernetes manifest
resource "kubernetes_manifest" "social_rating_flink_job" {
  # Read the manifest content from the YAML file relative to this main.tf
  manifest = yamldecode(file("${path.root}/../../social-rating-job.yaml"))

  # Explicitly depend on the Flink operator module being ready,
  # the JAR build/copy process completing, the namespace existing,
  # AND the dependent service modules completing their apply.
  depends_on = [
    module.flink_operator,
    null_resource.build_and_copy_flink_job,
    kubernetes_namespace.flink_jobs,
    module.scylladb, # Added dependency
    module.mongodb,
    module.dgraph,
    module.influxdb
  ]
}