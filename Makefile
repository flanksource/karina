
default: build
NAME:=karina

ifeq ($(VERSION),)
  VERSION_TAG=$(shell git describe --abbrev=0 --tags --exact-match 2>/dev/null || git describe --abbrev=0 --all)-$(shell date +"%Y%m%d%H%M%S")
else
  VERSION_TAG=$(VERSION)-$(shell date +"%Y%m%d%H%M%S")
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Image URL to use all building/pushing image targets
IMG ?= flanksource/karina:latest
# Produce CRDs with v1
CRD_OPTIONS ?= "crd:crdVersions=v1"

.PHONY: help
help:
	@cat docs/developer-guide/make-targets.md

.PHONY: release
release: linux darwin compress

.PHONY: build
build:
	go build -o ./.bin/$(NAME) -ldflags "-X \"main.version=$(VERSION_TAG)\""  main.go


.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X \"main.version=$(VERSION_TAG)\"" -o .bin/$(NAME)_linux-amd64
	cp .bin/$(NAME)_linux-amd64 .bin/$(NAME)
	GOOS=linux GOARCH=arm64 go build -ldflags "-X \"main.version=$(VERSION_TAG)\"" -o .bin/$(NAME)_linux-arm64

.PHONY: darwin
darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X \"main.version=$(VERSION_TAG)\""  -o .bin/$(NAME)_darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X \"main.version=$(VERSION_TAG)\"" -o .bin/$(NAME)_darwin-arm64

.PHONY: windows
windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X \"main.version=$(VERSION_TAG)\""  -o .bin/$(NAME).exe

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
	go run main.go docs api  pkg/types/inline.go pkg/types/config.go pkg/types/types.go pkg/types/nsx.go  > docs/reference/config.md
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
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=karina paths="./pkg/operator/..." output:rbac:artifacts:config=config/operator/rbac

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

