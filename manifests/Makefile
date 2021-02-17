ARGO_ROLLOUTS_VERSION := v0.10.2

TEMP_DIR := $(shell mktemp -d)

.PHONY: argo-rollouts
argo-rollouts: argo-rollouts-crds argo-rollouts-base

.PHONY: argo-rollouts-crds
argo-rollouts-crds:
	kustomize build "https://github.com/argoproj/argo-rollouts/manifests/crds?ref=$(ARGO_ROLLOUTS_VERSION)" > crd/argo-rollouts.yaml

define ARGO_ROLLOUTS_INSTALL_KUSTOMIZATION
bases:
  - github.com/argoproj/argo-rollouts/manifests/base?ref=$(ARGO_ROLLOUTS_VERSION)
  - github.com/argoproj/argo-rollouts/manifests/role?ref=$(ARGO_ROLLOUTS_VERSION)
resources:
  - https://raw.githubusercontent.com/argoproj/argo-rollouts/$(ARGO_ROLLOUTS_VERSION)/manifests/cluster-install/argo-rollouts-clusterrolebinding.yaml
images:
  - name: argoproj/argo-rollouts
    newTag: "{{.argoRollouts.version}}"
patchesStrategicMerge:
  - |-
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: argo-rollouts-clusterrolebinding
    subjects:
      - kind: ServiceAccount
        name: argo-rollouts
        namespace: platform-system
endef
export ARGO_ROLLOUTS_INSTALL_KUSTOMIZATION

.PHONY: argo-rollouts-base
argo-rollouts-base:
	echo "$$ARGO_ROLLOUTS_INSTALL_KUSTOMIZATION" > $(TEMP_DIR)/kustomization.yaml
	kustomize build $(TEMP_DIR) > argo-rollouts.yaml
	rm -rf $(TEMP_DIR)/*
