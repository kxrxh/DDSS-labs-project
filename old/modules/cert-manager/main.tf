terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.27"
    }
  }
}

# This resource installs cert-manager CRDs directly using kubectl
resource "null_resource" "install_cert_manager_crds" {
  triggers = {
    always_run = "${timestamp()}"
  }

  provisioner "local-exec" {
    command = <<-EOT
      echo "Installing cert-manager CRDs..."
      kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v${var.cert_manager_version}/cert-manager.crds.yaml || echo "Failed to install cert-manager CRDs, they might already exist"
      
      echo "Verifying cert-manager CRDs installation..."
      kubectl get crd | grep -q "certificates.cert-manager.io" || { echo "cert-manager CRDs not properly installed"; exit 1; }
      kubectl get crd | grep -q "issuers.cert-manager.io" || { echo "cert-manager CRDs not properly installed"; exit 1; }
      
      echo "cert-manager CRDs are properly installed"
    EOT
  }
}

# Install cert-manager using kubectl rather than Helm to avoid ownership conflicts
resource "null_resource" "install_cert_manager" {
  depends_on = [null_resource.install_cert_manager_crds]
  
  triggers = {
    always_run = "${timestamp()}"
  }

  provisioner "local-exec" {
    command = <<-EOT
      echo "Creating cert-manager namespace if it doesn't exist..."
      kubectl create namespace cert-manager --dry-run=client -o yaml | kubectl apply -f -
      
      echo "Checking if cert-manager is already installed..."
      if ! kubectl get deployment -n cert-manager cert-manager 2>/dev/null; then
        echo "Installing cert-manager v${var.cert_manager_version}..."
        kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v${var.cert_manager_version}/cert-manager.yaml
      else
        echo "cert-manager is already installed, skipping installation"
      fi
    EOT
  }
}

# Wait for cert-manager to be fully ready before proceeding
resource "null_resource" "wait_for_cert_manager" {
  depends_on = [null_resource.install_cert_manager]
  
  triggers = {
    always_run = "${timestamp()}"
  }

  provisioner "local-exec" {
    command = <<-EOT
      echo "Waiting for cert-manager deployments to be ready..."
      kubectl -n cert-manager wait --for=condition=available --timeout=${var.timeout}s deployment/cert-manager
      kubectl -n cert-manager wait --for=condition=available --timeout=${var.timeout}s deployment/cert-manager-cainjector
      kubectl -n cert-manager wait --for=condition=available --timeout=${var.timeout}s deployment/cert-manager-webhook
      
      echo "All cert-manager components are ready"
    EOT
  }
} 