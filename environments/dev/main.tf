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

resource "kubernetes_namespace" "analytics" {
  metadata {
    name = "dev-analytics"
  }
}

module "influxdb" {
  source = "../../modules/influxdb"

  namespace        = kubernetes_namespace.analytics.metadata[0].name
  release_name     = "dev-influxdb"
  # Let the module create the namespace if needed (redundant but safe)
  # create_namespace = true 
  persistence_storage_class = "local-path" # Use the same SC as redpanda
}

module "clickhouse" {
  source = "../../modules/clickhouse"

  # This module now deploys the main 'clickhouse' chart from Altinity,
  # which includes the operator as a dependency.

  # Release name for the main ClickHouse deployment
  operator_release_name     = "dev-clickhouse" # Renamed release

  # Namespace for BOTH the ClickHouse cluster AND the operator (if chart enables it)
  installation_namespace = kubernetes_namespace.analytics.metadata[0].name
  operator_create_namespace = true # Allow chart to create the installation namespace

  # installation_name is no longer needed (chart handles CHI naming)
  # installation_persistence_storage_class is now configured via chart values if needed
  # create_default_installation is removed as chart handles CHI
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

# Ensure the Flink job JAR is built and copied to the expected host path
resource "null_resource" "build_and_copy_flink_job" {
  # Trigger this resource when files in the Flink job source directory change
  triggers = {
    # Using timestamp of pom.xml as a proxy for changes
    # Consider a more robust method like hashing source files if needed
    source_code_hash = filemd5("${path.root}/../../social-rating-flink-job/pom.xml")
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
    module.clickhouse,
    module.redpanda,
    module.mongodb,
    module.dgraph,
    module.influxdb
  ]
}