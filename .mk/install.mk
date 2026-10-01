# ==== Installation targets ====

##@ Installation

.PHONY: install
install: $(KUSTOMIZE) ## Install temporal-resource-operator artifacts to the K8s cluster specified in ~/.kube/config.
	cd deploy/kustomize/resource-operator && "$(KUSTOMIZE)" edit set image resource-operator="$(BUILD_CONTAINER_IMAGE):$(BUILD_GIT_VERSION)"
	"$(KUSTOMIZE)" build deploy/kustomize/default | "$(KUBECTL)" apply -f -

.PHONY: uninstall
uninstall: $(KUSTOMIZE) ## Uninstall temporal-resource-operator artifacts from the K8s cluster specified in ~/.kube/config.
	"$(KUSTOMIZE)" build deploy/kustomize/default | "$(KUBECTL)" delete --ignore-not-found -f -
