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
  set {
    name  = "webhook.tls.auto"
    value = "true" # Explicitly set default, assumes cert-manager will be used.
  }
  set {
    name  = "webhook.tls.secretName"
    # This is the typical default pattern: <helm_release_name>-webhook-certs.
    # Setting it explicitly can help if the chart has issues determining it.
    # Ensure cert-manager is installed and can create this secret.
    value = "${var.scylla_operator_release_name}-webhook-certs"
  }

  # IMPORTANT: The following error indicates 'cert-manager' is not healthy or not found:
  # "failed calling webhook "webhook.cert-manager.io": ... service "cert-manager-webhook" not found"
  #
  # 1. VERIFY CERT-MANAGER: Ensure 'cert-manager' is correctly installed and its 'cert-manager-webhook'
  #    service is running in your cluster (typically in the 'cert-manager' namespace).
  #    Check 'kubectl get pods,svc -n cert-manager'.
  #
  # 2. TERRAFORM DEPENDENCY: If 'cert-manager' is also managed by this Terraform configuration
  #    (e.g., via another helm_release), you MUST ensure this Scylla operator release
  #    explicitly depends on the cert-manager release to ensure it's fully operational first.
  #
  # Example if cert-manager is a Helm release named 'cert_manager' in the same Terraform config:
  # depends_on = [
  #   helm_release.cert_manager # Replace 'helm_release.cert_manager' with your actual cert-manager resource
  # ]
  #
  # If cert-manager is installed manually or by other means, ensure it is healthy before applying this.
  depends_on = [
    #kubernetes_namespace.scylla_db_ns # Already a dependency for the cluster resource
    # Example: helm_release.cert_manager_resource_name  <-- UNCOMMENT AND SET THIS if cert-manager is managed by Terraform
    var.cert_manager_helm_release
  ]
}

resource "kubernetes_namespace" "scylla_db_ns" {
  metadata {
    name = var.scylla_namespace
  }

  lifecycle {
    ignore_changes = [
      metadata[0].labels,
      metadata[0].annotations
    ]
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
      version       = var.scylla_version # Using variable instead of hardcoded version
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
    kubernetes_namespace.scylla_db_ns,
    var.cert_manager_helm_release # Also ensure cluster depends on cert-manager if operator does
  ]
}

resource "kubernetes_job_v1" "keyspace_creation_job" {
  metadata {
    name      = "${var.scylla_cluster_name}-keyspace-social-rating-${substr(sha256("${var.scylla_cluster_name}-${var.scylla_default_keyspace_name}-${timestamp()}"), 0, 8)}"
    namespace = var.scylla_namespace
  }

  spec {
    template {
      metadata {
        labels = {
          app = "${var.scylla_cluster_name}-keyspace-creator"
        }
      }
      spec {
        restart_policy = "OnFailure"
        container {
          name  = "cqlsh-runner"
          image = "scylladb/scylla:${var.scylla_version}"
          command = ["/bin/sh", "-c"]
          args = [
            <<-EOT
              echo "Waiting for ScyllaDB to be ready..."
              for i in $(seq 1 30); do
                if cqlsh -e 'DESCRIBE KEYSPACES' ${var.scylla_cluster_name}-client.${var.scylla_namespace}.svc.cluster.local 9042; then
                  echo "ScyllaDB is ready!"
                  cqlsh -e "CREATE KEYSPACE IF NOT EXISTS ${var.scylla_default_keyspace_name} WITH replication = {'class': 'SimpleStrategy', 'replication_factor': ${var.scylla_replication_factor}};" ${var.scylla_cluster_name}-client.${var.scylla_namespace}.svc.cluster.local 9042
                  exit 0
                fi
                echo "Attempt $i: ScyllaDB not ready yet, waiting..."
                sleep 20
              done
              echo "Failed to connect to ScyllaDB after 30 attempts"
              exit 1
            EOT
          ]
        }
      }
    }
    backoff_limit = 4
    ttl_seconds_after_finished = 300
  }

  timeouts {
    create = "15m"
    update = "15m"
  }

  wait_for_completion = true

  lifecycle {
    replace_triggered_by = [
      kubernetes_manifest.scylla_cluster
    ]
  }

  depends_on = [
    kubernetes_manifest.scylla_cluster
  ]
}

resource "kubernetes_job_v1" "table_creation_job" {
  metadata {
    name      = "${var.scylla_cluster_name}-tables-creation-${substr(sha256("${var.scylla_cluster_name}-${var.scylla_default_keyspace_name}-${timestamp()}"), 0, 8)}"
    namespace = var.scylla_namespace
  }

  spec {
    template {
      metadata {
        labels = {
          app = "${var.scylla_cluster_name}-table-creator"
        }
      }
      spec {
        restart_policy = "OnFailure"
        container {
          name  = "cqlsh-runner"
          image = "scylladb/scylla:${var.scylla_version}"
          command = ["/bin/sh", "-c"]
          args = [
            <<-EOT
              echo "Creating tables in ScyllaDB..."
              
              # Create events_archive table
              cqlsh -e "USE ${var.scylla_default_keyspace_name}; CREATE TABLE IF NOT EXISTS events_archive (citizen_id text, event_id text, event_type text, event_subtype text, source_system text, timestamp timestamp, location_city text, location_region text, location_country text, confidence double, score_change double, new_score double, previous_score double, applied_rules list<text>, PRIMARY KEY (citizen_id, timestamp, event_id)) WITH CLUSTERING ORDER BY (timestamp DESC, event_id ASC) AND default_time_to_live = 157680000;" ${var.scylla_cluster_name}-client.${var.scylla_namespace}.svc.cluster.local 9042
              
              if [ $? -ne 0 ]; then
                echo "Failed to create events_archive table"
                exit 1
              fi
              
              # Create rule_application_state table
              cqlsh -e "USE ${var.scylla_default_keyspace_name}; CREATE TABLE IF NOT EXISTS rule_application_state (citizen_id text, rule_id text, last_applied timestamp, cooldown_until timestamp, application_count int, PRIMARY KEY ((citizen_id), rule_id));" ${var.scylla_cluster_name}-client.${var.scylla_namespace}.svc.cluster.local 9042
              
              if [ $? -eq 0 ]; then
                echo "Tables created successfully!"
              else
                echo "Failed to create rule_application_state table"
                exit 1
              fi
            EOT
          ]
        }
      }
    }
    backoff_limit = 4
    ttl_seconds_after_finished = 300
  }

  timeouts {
    create = "15m"
    update = "15m"
  }

  wait_for_completion = true

  lifecycle {
    replace_triggered_by = [
      kubernetes_job_v1.keyspace_creation_job
    ]
  }

  depends_on = [
    kubernetes_job_v1.keyspace_creation_job
  ]
} 