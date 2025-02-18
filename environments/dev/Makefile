init:
	tofu init

plan:
	tofu plan

apply:
	tofu apply --auto-approve


destroy:
	tofu destroy

destroy-namespace:
	kubectl delete namespace dgraph-dev || true

destroy-all-namespaces:
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
		xargs kubectl delete namespace; \
	fi

.PHONY: init plan apply destroy destroy-namespace destroy-all-namespaces
