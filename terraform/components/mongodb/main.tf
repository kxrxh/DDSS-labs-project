resource "kubernetes_namespace" "mongodb_ns" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_service" "mongodb_headless" {
  metadata {
    name      = "mongodb-service"
    # namespace = kubernetes_namespace.mongodb.metadata[0].name
    namespace = var.namespace # Use the passed-in namespace directly
    labels = {
      app = "mongodb"
    }
  }
  spec {
    cluster_ip = "None"
    port {
      port        = 27017
      target_port = 27017
    }
    selector = {
      app = "mongodb"
    }
  }
}

resource "kubernetes_stateful_set" "mongodb" {
  metadata {
    name      = "mongodb"
    # namespace = kubernetes_namespace.mongodb.metadata[0].name
    namespace = var.namespace # Use the passed-in namespace directly
  }

  # lifecycle {
  #   replace_triggered_by = [
  #     kubernetes_namespace.mongodb # No longer needed as ns is managed outside
  #   ]
  # }

  spec {
    service_name = kubernetes_service.mongodb_headless.metadata[0].name
    replicas     = var.replicas

    selector {
      match_labels = {
        app = "mongodb"
      }
    }

    template {
      metadata {
        labels = {
          app = "mongodb"
        }
      }

      spec {
        container {
          name  = "mongodb"
          image = var.mongo_image

          port {
            container_port = 27017
          }

          volume_mount {
            name       = "mongodb-data"
            mount_path = "/data/db"
          }
        }
      }
    }

    volume_claim_template {
      metadata {
        name = "mongodb-data"
      }

      spec {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = var.storage_size
          }
        }
      }
    }
  }
}

# Job to initialize MongoDB with required system configuration
resource "kubernetes_job_v1" "mongodb_init_job" {
  metadata {
    name      = "mongodb-init-social-rating"
    namespace = var.namespace
  }

  spec {
    template {
      metadata {
        labels = {
          app = "mongodb-initializer"
        }
      }
      spec {
        restart_policy = "OnFailure"
        container {
          name  = "mongodb-init"
          image = "mongo:7"
          command = ["/bin/sh", "-c"]
          args = [
            <<-EOT
              echo "Waiting for MongoDB to be ready..."
              for i in $(seq 1 30); do
                if mongosh --host mongodb-service.${var.namespace}.svc.cluster.local:27017 --eval "db.adminCommand('ping')" 2>/dev/null; then
                  echo "MongoDB is ready!"
                  break
                fi
                echo "Attempt $i: MongoDB not ready yet, waiting..."
                sleep 10
              done

              echo "Initializing MongoDB with system configuration..."
              mongosh --host mongodb-service.${var.namespace}.svc.cluster.local:27017 <<'MONGO_SCRIPT'
              use social_rating

              // Create system configuration
              db.system_configuration.replaceOne(
                { "_id": "default" },
                {
                  "_id": "default",
                  "baseline_score": parseFloat("500.0"),
                  "max_score": parseFloat("1000.0"),
                  "min_score": parseFloat("0.0"),
                  "tier_definitions": {
                    "excellent": {
                      "min_score": parseFloat("800.0"),
                      "max_score": parseFloat("1000.0"),
                      "name": "Excellent",
                      "color": "#00ff00",
                      "description": "Exemplary citizens with high trust score"
                    },
                    "good": {
                      "min_score": parseFloat("600.0"),
                      "max_score": parseFloat("799.0"),
                      "name": "Good",
                      "color": "#90ee90",
                      "description": "Good citizens with adequate trust score"
                    },
                    "average": {
                      "min_score": parseFloat("400.0"),
                      "max_score": parseFloat("599.0"),
                      "name": "Average",
                      "color": "#ffff00",
                      "description": "Average citizens requiring some attention"
                    },
                    "poor": {
                      "min_score": parseFloat("200.0"),
                      "max_score": parseFloat("399.0"),
                      "name": "Poor",
                      "color": "#ffa500",
                      "description": "Citizens with concerning behavior patterns"
                    },
                    "very_poor": {
                      "min_score": parseFloat("0.0"),
                      "max_score": parseFloat("199.0"),
                      "name": "Very Poor",
                      "color": "#ff0000",
                      "description": "High-risk citizens requiring intervention"
                    }
                  },
                  "active_rule_set": "default",
                  "score_decay_policy": {
                    "enabled": false,
                    "rate": 0.1,
                    "interval": "daily"
                  },
                  "updated_at": new Date(),
                  "version": 1
                },
                { upsert: true }
              )

              // Create some basic scoring rules
              db.scoring_rule_definitions.insertMany([
                {
                  "_id": "financial_positive",
                  "name": "Financial Positive Behavior",
                  "event_type": "financial",
                  "conditions": {
                    "event_subtype": "payment_completed",
                    "amount": { "$gt": 0 }
                  },
                  "points": parseFloat("10.0"),
                  "multiplier": parseFloat("1.0"),
                  "status": "active",
                  "valid_from": new Date(),
                  "valid_to": null,
                  "created_at": new Date(),
                  "updated_at": new Date(),
                  "description": "Reward for completed payments"
                },
                {
                  "_id": "financial_negative",
                  "name": "Financial Negative Behavior",
                  "event_type": "financial",
                  "conditions": {
                    "event_subtype": "payment_failed"
                  },
                  "points": parseFloat("-15.0"),
                  "multiplier": parseFloat("1.0"),
                  "status": "active",
                  "valid_from": new Date(),
                  "valid_to": null,
                  "created_at": new Date(),
                  "updated_at": new Date(),
                  "description": "Penalty for failed payments"
                },
                {
                  "_id": "social_positive",
                  "name": "Social Positive Behavior",
                  "event_type": "social",
                  "conditions": {
                    "event_subtype": "volunteering"
                  },
                  "points": parseFloat("20.0"),
                  "multiplier": parseFloat("1.2"),
                  "status": "active",
                  "valid_from": new Date(),
                  "valid_to": null,
                  "created_at": new Date(),
                  "updated_at": new Date(),
                  "description": "Reward for volunteering activities"
                },
                {
                  "_id": "legal_violation",
                  "name": "Legal Violation",
                  "event_type": "legal",
                  "conditions": {
                    "event_subtype": "violation"
                  },
                  "points": parseFloat("-50.0"),
                  "multiplier": parseFloat("1.5"),
                  "status": "active",
                  "valid_from": new Date(),
                  "valid_to": null,
                  "created_at": new Date(),
                  "updated_at": new Date(),
                  "description": "Penalty for legal violations"
                },
                {
                  "_id": "relationship_positive",
                  "name": "Positive Relationship Impact",
                  "event_type": "relationship",
                  "conditions": {
                    "is_positive": true
                  },
                  "points": parseFloat("5.0"),
                  "multiplier": parseFloat("1.0"),
                  "status": "active",
                  "valid_from": new Date(),
                  "valid_to": null,
                  "created_at": new Date(),
                  "updated_at": new Date(),
                  "description": "Reward for positive social interactions that affect relationships"
                },
                {
                  "_id": "relationship_negative",
                  "name": "Negative Relationship Impact",
                  "event_type": "relationship",
                  "conditions": {
                    "is_positive": false
                  },
                  "points": parseFloat("-8.0"),
                  "multiplier": parseFloat("1.0"),
                  "status": "active",
                  "valid_from": new Date(),
                  "valid_to": null,
                  "created_at": new Date(),
                  "updated_at": new Date(),
                  "description": "Penalty for negative social interactions that affect relationships"
                }
              ])

              // Create a few sample citizens for testing
              db.citizens.insertMany([
                {
                  "_id": "citizen_001",
                  "age": 30,
                  "score": parseFloat("500.0"),
                  "type": "regular",
                  "status": "active",
                  "tier": "average",
                  "created_at": new Date(),
                  "last_updated": new Date(),
                  "last_score_at": new Date()
                },
                {
                  "_id": "citizen_002",
                  "age": 45,
                  "score": parseFloat("750.0"),
                  "type": "regular",
                  "status": "active",
                  "tier": "good",
                  "created_at": new Date(),
                  "last_updated": new Date(),
                  "last_score_at": new Date()
                },
                {
                  "_id": "citizen_003",
                  "age": 25,
                  "score": parseFloat("300.0"),
                  "type": "regular",
                  "status": "active",
                  "tier": "poor",
                  "created_at": new Date(),
                  "last_updated": new Date(),
                  "last_score_at": new Date()
                }
              ])

              print("MongoDB initialization completed successfully!")
              MONGO_SCRIPT

              if [ $? -eq 0 ]; then
                echo "MongoDB initialization completed successfully!"
              else
                echo "Failed to initialize MongoDB"
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

  depends_on = [
    kubernetes_stateful_set.mongodb
  ]
}