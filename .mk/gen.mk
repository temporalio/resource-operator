# ==== Code generation targets ====

CODEGEN_DIR=$(BIN_DIR)/code-generator
KUBE_CODEGEN_PATH=$(CODEGEN_DIR)/kube_codegen.sh
$(KUBE_CODEGEN_PATH): | $(BIN_DIR)
	@echo -n "installing k8s.io/code-generator ... "
	@git clone https://github.com/kubernetes/code-generator.git $(CODEGEN_DIR) >/dev/null 2>&1
	@echo "ok."

KUSTOMIZE ?= $(BIN_DIR)/kustomize
KUSTOMIZE_VERSION ?= v5.8.2
$(KUSTOMIZE): | $(BIN_DIR)
	@echo -n "installing kustomize@$(KUSTOMIZE_VERSION) ... "
	@$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))
	@echo "ok."

CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.21.0
$(CONTROLLER_GEN): | $(BIN_DIR)
	@echo -n "installing controller-tools@$(CONTROLLER_TOOLS_VERSION) ... "
	@$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))
	@echo "ok."

##@ Code generation

.PHONY: gen-all
gen-all: gen-kube-helpers gen-kube-register gen-kube-manifests ## Run all code generation targets.

.PHONY: gen-ensure
gen-ensure: $(KUSTOMIZE) $(CONTROLLER_GEN) $(KUBE_CODEGEN_PATH) ## Download all code generation utilities locally if necessary.

.PHONY: gen-kube-helpers
KUBE_CODEGEN_HELPERS_PATH=$(SCRIPTS_DIR)/kube-codegen-helpers.sh
gen-kube-helpers: $(KUBE_CODEGEN_PATH) ## Generate API deepcopy and validation Go code
	@echo -n "generating API deepcopy and validation Go code ... "
	@$(KUBE_CODEGEN_HELPERS_PATH) >/dev/null 2>&1
	@echo "ok."

.PHONY: gen-kube-register
KUBE_CODEGEN_REGISTER_PATH=$(SCRIPTS_DIR)/kube-codegen-register.sh
gen-kube-register: $(KUBE_CODEGEN_PATH) ## Generate API scheme and groupversion Go code
	@echo -n "generating API scheme and groupversion Go code ... "
	@$(KUBE_CODEGEN_REGISTER_PATH) >/dev/null 2>&1
	@echo "ok."

.PHONY: gen-kube-manifests
CONTROLLER_ROLE_NAME ?= temporal-resource-operator
# Note that the option maxDescLen=0 was added in the default scaffold in
# order to sort out the issue Too long: must have at most 262144 bytes. By
# using kubectl apply to create / update resources an annotation is created
# by K8s API to store the latest version of the resource (
# kubectl.kubernetes.io/last-applied-configuration).  However, it has a
# size limit and if the CRD is too big with so many long descriptions as
# this one it will cause the failure.
gen-kube-manifests: $(CONTROLLER_GEN) ## Generate WebhookConfiguration, ClusterRole objects.
	@echo -n "generating Kubernetes manifests ... "
	@$(CONTROLLER_GEN) rbac:roleName=$(CONTROLLER_ROLE_NAME) \
		crd:maxDescLen=0 webhook paths="./api/...;./controller/..." \
		output:crd:dir=deploy/kustomize/crd/bases \
		output:rbac:dir=deploy/kustomize/rbac
	@echo "ok."
