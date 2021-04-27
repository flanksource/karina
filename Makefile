
default: build
NAME:=karina

ifeq ($(VERSION),)
VERSION=$(shell git describe --tags  --long)-$(shell date +"%Y%m%d%H%M%S")
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Image URL to use all building/pushing image targets
IMG ?= flanksource/karina:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

.PHONY: help
help:
	@cat docs/developer-guide/make-targets.md

.PHONY: release
release: linux darwin compress

.PHONY: build
build:
	go build -o ./.bin/$(NAME) -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: linux
linux:
	GOOS=linux go build -o ./.bin/$(NAME) -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: darwin
darwin:
	GOOS=darwin go build -o ./.bin/$(NAME)_osx -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: compress
compress:
	upx -5 ./.bin/*

.PHONY: install
install:
	cp ./.bin/$(NAME) /usr/local/bin/

.PHONY: docker
docker:
	docker build ./ -t $(NAME)

.PHONY: docker-fast
docker-fast: linux-static
	docker build ./ -t $(IMG) -f Dockerfile.fast

.PHONY: serve-docs
serve-docs:
	cd docs && make serve

.PHONY: build-api-docs
build-api-docs:
	go run main.go docs api  pkg/types/config.go pkg/types/types.go pkg/types/nsx.go  > docs/reference/config.md
	go run main.go docs cli "docs/cli"

.PHONY: build-docs
build-docs:
	cd docs && make  build


.PHONY: deploy-docs
deploy-docs:
	cd docs && make deploy

.PHONY: lint
lint: build
	golangci-lint run --verbose --print-resources-usage
	go run test/linter/main.go

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/api/operator/..."
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/types/..."
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/api/calico/..."

static: manifests
	mkdir -p config/deploy
	cd config/operator/manager && kustomize edit set image controller=${IMG}
	kustomize build config/crd > config/deploy/crd.yml
	kustomize build config/operator/default > config/deploy/operator.yml

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/operator/manager && kustomize edit set image controller=${IMG}
	kustomize build config/operator/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) webhook paths="./pkg/api/operator/..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager paths="./pkg/operator/..." output:rbac:artifacts:config=config/operator/rbac

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

