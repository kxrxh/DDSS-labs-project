resource "helm_release" "influxdb" {
  name             = var.release_name
  namespace        = var.namespace
  create_namespace = var.create_namespace

  repository       = "https://helm.influxdata.com/"
  chart            = "influxdb2"
  version          = var.chart_version

  values = [
    yamlencode({
      persistence = {
        enabled = var.persistence_enabled
        size    = var.persistence_size
        # Set storageClassName only if specified, otherwise let Helm/K8s use default
        storageClassName = var.persistence_storage_class != null ? var.persistence_storage_class : ""
      }
      # Add other InfluxDB configurations here if needed
      # Refer to the chart's values: https://github.com/influxdata/helm-charts/blob/master/charts/influxdb2/values.yaml
    })
  ]

  timeout = 600 # 10 minutes

  # Optional: Set up admin user/token if needed via values (refer to chart docs)
}

# Note: This module deploys InfluxDB 2.x.
# You will likely need to configure buckets, tokens, etc. post-deployment
# either manually or using the InfluxDB API/CLI or potentially Terraform providers. 