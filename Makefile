# Variables
FLINK_JOB_DIR := social-rating-flink-job
TARGET_DIR := $(FLINK_JOB_DIR)/target
SOURCE_JAR_NAME := social-rating-flink-job-1.0-SNAPSHOT.jar
SOURCE_JAR_PATH := $(TARGET_DIR)/$(SOURCE_JAR_NAME)

HOST_JAR_DIR := /tmp/flink-jars
DEST_JAR_NAME := social-rating-flink-job.jar
DEST_JAR_PATH := $(HOST_JAR_DIR)/$(DEST_JAR_NAME)

K8S_NAMESPACE := dev-flink-jobs
K8S_MANIFEST := social-rating-job.yaml

# Phony targets (not associated with files)
.PHONY: all build copy deploy clean

# Default target
all: deploy

# Build the Flink job JAR
build:
	@echo "+++ Building Flink job JAR..."
	@cd $(FLINK_JOB_DIR) && mvn clean package
	@echo "+++ Build complete."

# Copy the JAR to the host path directory
copy: build
	@echo "+++ Copying JAR to $(HOST_JAR_DIR)..."
	@mkdir -p $(HOST_JAR_DIR)
	@cp $(SOURCE_JAR_PATH) $(DEST_JAR_PATH)
	@echo "+++ Copy complete: $(DEST_JAR_PATH)"

# Deploy the Flink job to Kubernetes
deploy: copy
	@echo "+++ Deploying Flink job to Kubernetes namespace $(K8S_NAMESPACE)..."
	@kubectl create namespace $(K8S_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	@kubectl apply -f $(K8S_MANIFEST)
	@echo "+++ Deployment initiated. Monitor status with: kubectl get pods -n $(K8S_NAMESPACE)"

# Clean up build artifacts and copied JAR
clean:
	@echo "+++ Cleaning build artifacts and copied JAR..."
	@cd $(FLINK_JOB_DIR) && mvn clean
	@rm -f $(DEST_JAR_PATH)
	@echo "+++ Clean complete." 