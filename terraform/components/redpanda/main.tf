resource "helm_release" "redpanda" {
  name       = var.redpanda_name
  repository = var.redpanda_repository
  chart      = var.redpanda_chart_name
  version    = var.redpanda_chart_version
  namespace  = var.redpanda_namespace
  create_namespace = true

  values = [
    <<-EOT
    # Add any specific Redpanda Helm values here
    # For example, to match the Ansible deployment defaults:
    # enterprise:
    #   license: "" # Add your enterprise license if you have one
    # external:
    #   domain: "local.redpanda.com" # Example domain
    # cluster:
    #   # As per the Ansible defaults, we'll keep these fairly standard.
    #   # Refer to Redpanda Helm chart documentation for more options.
    tls:
      enabled: false # Corresponds to tls=false in Ansible vars
    # developerMode: false # Corresponds to redpanda_mode=production (developer_mode: false in JSONDATA example)
    # See https://docs.redpanda.com/current/deploy/deployment-option/self-hosted/kubernetes/production/customize-helm-chart/
    # for more details on Helm chart configuration.

    # Default ports from Ansible variables:
    # adminApi:
    #   port: 9644
    # kafkaApi:
    #   port: 9092
    # rpcApi:
    #   port: 33145
    # schemaRegistry:
    #   port: 8081
    #   enabled: true # Assuming it's enabled if port is specified

    # Default ports and other configurations are generally fine unless specific needs arise.
    # For a single-node cluster (like Minikube, Kind, OrbStack),
    # we need to adjust the default 'hard' anti-affinity to 'soft' or disable it.
    statefulset:
      replicas: 1 # Changed from default of 3
      podAntiAffinity:
        type: soft # Was 'hard', preventing scheduling on a single node

    # Example of overriding resources for Redpanda statefulset
    resources:
      requests:
        cpu: 1 # Using the default from your node output
        memory: "2500Mi" # Reduced from 2560Mi
      limits:
        cpu: 1 # Using the default from your node output
        memory: "2500Mi" # Reduced from 2560Mi

    EOT
  ]

  set {
    name  = "console.enabled"
    value = var.redpanda_console_enabled
  }

  # Add more 'set' blocks for other specific values if needed
  # Example from Ansible (though these might be structured differently in Helm):
  # set {
  #   name = "cluster.rackAwareness.enabled" # Fictional path, check chart docs
  #   value = var.redpanda_rack_awareness_enabled
  # }
  # set {
  #   name = "storage.tiered.bucket" # Fictional path, check chart docs
  #   value = var.redpanda_tiered_storage_bucket_name
  #   type  = "string" # Specify type if it's not a string by default or to ensure it is treated as string
  # }
} 