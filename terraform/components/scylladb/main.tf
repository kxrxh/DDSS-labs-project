resource "helm_release" "scylla_operator" {
  name             = var.scylla_operator_release_name
  repository       = var.scylla_operator_repository
  chart            = var.scylla_operator_chart_name
  version          = var.scylla_operator_chart_version
  namespace        = var.scylla_operator_namespace
  create_namespace = var.scylla_operator_create_namespace

  # ScyllaDB Operator doesn't usually require many value overrides for a basic setup.
  # Refer to the official ScyllaDB Operator Helm chart documentation for available options.
  # Example: https://operator.docs.scylladb.com/stable/helm.html
  values = [
    <<-EOT
    # Add any specific ScyllaDB Operator Helm values here if needed
    EOT
  ]

  set {
    name  = "crds.create"
    value = "true" # Explicitly set, though it's usually the default
  }

  depends_on = [] # Add dependencies if any, e.g., cert-manager if TLS is used for operator webhooks
}

resource "kubernetes_namespace" "scylla_db_ns" {
  metadata {
    name = var.scylla_namespace
  }
}

# This resource creates a ScyllaDBCluster custom resource. 
# The ScyllaDB Operator, once installed by the helm_release above, will detect this CR 
# and provision a ScyllaDB cluster accordingly.
resource "kubernetes_manifest" "scylla_cluster" {
  manifest = {
    apiVersion = "scylla.scylladb.com/v1"
    kind       = "ScyllaCluster"
    metadata = {
      name      = var.scylla_cluster_name
      namespace = var.scylla_namespace
    }
    spec = {
      version       = "5.4.6" # Added ScyllaDB version
      developerMode = true    # For easier setup in non-production
      agentVersion  = var.scylla_agent_version
      datacenter = {
        name  = var.scylla_cluster_datacenter_name
        racks = [
          {
            name    = var.scylla_cluster_rack_name
            members = var.scylla_cluster_members
            storage = {
              capacity = "10Gi" # Define storage capacity per member
            }
            resources = {
              requests = {
                cpu    = var.scylla_cluster_cpus
                memory = var.scylla_cluster_memory
              }
              limits = {
                cpu    = var.scylla_cluster_cpus
                memory = var.scylla_cluster_memory
              }
            }
            # For local setups like Minikube/Kind/OrbStack, you might need to allow scheduling on same nodes
            # if your Kubernetes cluster is small.
            # placement = {
            #   tolerations = [
            #     {
            #       key = "key",
            #       operator = "Equal",
            #       value = "value",
            #       effect = "NoSchedule"
            #     }
            #   ]
            # }
          }
        ]
      }
      # repairs = [
      #   {
      #     name = "weekly-repair",
      #     schedule = "0 0 * * 0", # Every Sunday at midnight
      #     intensity = "1",
      #     parallel = 1,
      #     smallTableThreshold = "1Gi",
      #     dc = [var.scylla_cluster_datacenter_name]
      #   }
      # ]
      # backups = [] # Define backup schedules if needed
    }
  }

  depends_on = [
    helm_release.scylla_operator,
    kubernetes_namespace.scylla_db_ns
  ]
} 