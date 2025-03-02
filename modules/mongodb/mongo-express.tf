resource "kubernetes_deployment" "mongo_express" {
  metadata {
    name      = "mongo-express"
    namespace = kubernetes_namespace.mongodb.metadata[0].name
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "mongo-express"
      }
    }

    template {
      metadata {
        labels = {
          app = "mongo-express"
        }
      }

      spec {
        container {
          name  = "mongo-express"
          image = "mongo-express:latest"

          port {
            container_port = 8081
          }

          env {
            name  = "ME_CONFIG_MONGODB_SERVER"
            value = kubernetes_service.mongodb_headless.metadata[0].name
          }

          env {
            name  = "ME_CONFIG_MONGODB_PORT"
            value = "27017"
          }

          env {
            name  = "ME_CONFIG_BASICAUTH_USERNAME"
            value = "admin"
          }

          env {
            name  = "ME_CONFIG_BASICAUTH_PASSWORD"
            value = "pass"  # For production, consider using a kubernetes_secret resource
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "mongo_express" {
  metadata {
    name      = "mongo-express-service"
    namespace = kubernetes_namespace.mongodb.metadata[0].name
  }

  spec {
    selector = {
      app = "mongo-express"
    }

    port {
      port        = 8081
      target_port = 8081
    }

    type = "NodePort"
  }
} 