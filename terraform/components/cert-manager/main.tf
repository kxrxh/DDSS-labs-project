variable "cert_manager_namespace" {
  description = "Namespace for cert-manager."
  type        = string
  default     = "cert-manager"
}

variable "cert_manager_chart_version" {
  description = "Version of the cert-manager Helm chart to deploy."
  type        = string
  default     = "v1.14.5" # Check for the latest stable version
}

resource "kubernetes_namespace" "cert_manager_ns" {
  metadata {
    name = var.cert_manager_namespace
  }
}

resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = var.cert_manager_chart_version
  namespace  = kubernetes_namespace.cert_manager_ns.metadata[0].name

  # CRDs are not installed by default with Helm 3+ and must be explicitly enabled.
  # Or, you can install them separately using kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/YOUR_CHART_VERSION/cert-manager.crds.yaml
  set {
    name  = "installCRDs"
    value = "true"
  }

  # Other common values you might consider:
  # set {
  #   name  = "prometheus.enabled"
  #   value = "false" # Set to true if you use Prometheus for monitoring
  # }
  # set {
  #   name = "webhook.hostNetwork" # Needed for some CNI configurations like Calico with GKE.
  #   value = "true"
  # }
  # set {
  #    name = "webhook.securePort"
  #    value = "10250" # Default, change if needed
  # }

  depends_on = [
    kubernetes_namespace.cert_manager_ns
  ]

  # It's good practice to wait for the deployment to be ready,
  # especially if other resources depend on cert-manager webhook.
  atomic         = true
  cleanup_on_fail = true
  timeout        = 300 # 5 minutes
}

# You might also want to define some default ClusterIssuers here once cert-manager is up.
# For example, a self-signed one for testing, or Let's Encrypt for public domains.
# resource "kubernetes_manifest" "letsencrypt_staging_clusterissuer" {
#   provider = kubernetes-alpha # If using kubernetes_manifest
#   manifest = {
#     "apiVersion" = "cert-manager.io/v1"
#     "kind"       = "ClusterIssuer"
#     "metadata"   = {
#       "name" = "letsencrypt-staging"
#     }
#     "spec" = {
#       "acme" = {
#         "server" = "https://acme-staging-v02.api.letsencrypt.org/directory"
#         "email"  = "your-email@example.com" # Change this
#         "privateKeySecretRef" = {
#           "name" = "letsencrypt-staging-issuer-account-key"
#         }
#         "solvers" = [
#           {
#             "http01" = {
#               "ingress" = {
#                 "class" = "nginx" # Or your ingress controller class
#               }
#             }
#           }
#         ]
#       }
#     }
#   }
#   depends_on = [helm_release.cert_manager]
# } 