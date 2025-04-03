init:
	tofu init

plan:
	tofu plan

apply:
	tofu apply --auto-approve


destroy:
	@echo "Uninstalling Helm releases to allow operators to clean up..."
	helm uninstall dev-clickhouse -n dev-analytics || true
	helm uninstall dev-influxdb -n dev-analytics || true
	helm uninstall dev-flink-operator -n dev-flink-operator || true
	helm uninstall dev-redpanda -n dev-streaming || true
	helm uninstall dev-mongodb -n dev-db || true # Assuming release name/namespace
	helm uninstall dev-dgraph -n dev-db || true # Assuming release name/namespace
	@echo "Waiting 15 seconds for cleanup..."
	sleep 15
	@echo "Running tofu destroy..."
	tofu destroy --auto-approve

destroy-all-namespaces:
	@echo "WARNING: This command is deprecated. Use 'make clean-all' instead."
	@echo "WARNING: This will delete ALL namespaces except those listed below."
	@echo "Make sure you understand the consequences before proceeding."
	@echo "Namespaces to be preserved:"
	@echo "  - kube-system"
	@echo "  - kube-public"
	@echo "  - kube-node-lease"
	@echo "  - default"
	@read -p "Are you sure you want to continue? (y/N) " -r ; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		kubectl get namespaces --no-headers -o custom-columns=":metadata.name" | \
		grep -v -E '^(kube-system|kube-public|kube-node-lease|default)' | \
		xargs -r kubectl delete namespace --force --grace-period=0; \
	fi

# Target to run destroy and then aggressively clean up stuck namespaces
clean-all:
	@echo "Attempting to remove finalizers from resources in terminating namespaces..."
	@# Get namespaces stuck in Terminating state (excluding system ones)
	@terminating_ns=$$(kubectl get ns --field-selector status.phase=Terminating --no-headers -o custom-columns=":metadata.name" | grep -v -E '^(kube-system|kube-public|kube-node-lease|default)' || true); \
	if [ -n "$$terminating_ns" ]; then \
		for ns in $$terminating_ns; do \
			echo "Processing terminating namespace: $$ns"; \
			echo "  Attempting to patch known resources (pods, pvcs, chi, flinkdeployments)..."; \
			for resource_type in pods pvc clickhouseinstallations.clickhouse.altinity.com flinkdeployments.flink.apache.org; do \
				echo "    Checking resource type: $$resource_type in $$ns..."; \
				resources=$$(kubectl get $$resource_type -n $$ns --ignore-not-found -o jsonpath='{range .items[?(@.metadata.finalizers)]}{.metadata.name}{"\n"}{end}' || true); \
				if [ -n "$$resources" ]; then \
					for resource_name in $$resources; do \
						echo "      Patching $$resource_type/$$resource_name in $$ns to remove finalizers..."; \
						kubectl patch $$resource_type $$resource_name -n $$ns --type merge -p '{"metadata":{"finalizers":null}}' || echo "      Patch failed for $$resource_type/$$resource_name (might already be deleted or other issue)"; \
					done; \
				fi; \
			done; \
			echo "  Finished patching attempts for $$ns."; \
		done; \
		echo "Waiting 10 seconds after patching finalizers..."; \
		sleep 10; \
	else \
		echo "No namespaces found in Terminating state."; \
	fi

	@echo "Proceeding to force-delete remaining non-system namespaces..."
	@echo "WARNING: This will attempt to delete ALL namespaces except those listed below."
	@echo "Make sure you understand the consequences before proceeding."
	@echo "Namespaces to be preserved:"
	@echo "  - kube-system"
	@echo "  - kube-public"
	@echo "  - kube-node-lease"
	@echo "  - default"
	@# No prompt here, assume if they run clean-all they want the cleanup
	@# Force delete remaining non-system namespaces
	@kubectl get namespaces --no-headers -o custom-columns=":metadata.name" | \
	grep -v -E '^(kube-system|kube-public|kube-node-lease|default)' | \
	xargs -r kubectl delete namespace --force --grace-period=0 --ignore-not-found || echo "Namespace deletion command finished (ignore errors if namespaces already gone)";
	@echo "Namespace force-deletion initiated (or none found)."

.PHONY: init plan apply destroy destroy-namespace destroy-all-namespaces clean-all
