
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.25.0
K3S_VERSION =v$(ENVTEST_K8S_VERSION)-k3s1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName='PLACEHOLDER' crd webhook paths="./..." output:rbac:artifacts:config=helm/templates output:crd:artifacts:config=helm/crds/
	sed -i '' 's/PLACEHOLDER/\{\{ include \"helm.fullname\" \. \}\}-controller/g' helm/templates/role.yaml

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: dev
dev: create_cluster deploy install_example

.PHONY: install_example
install_example: ## install an example in the current cluster
	KO_DOCKER_REPO=kxds-registry.localhost:5000 ko apply -f ./example/k8s/echo-server
	KO_DOCKER_REPO=kxds-registry.localhost:5000 ko apply -f ./example/k8s/echo-client

.PHONY: create_cluster
create_cluster: ## run a local k3d cluster
	k3d cluster create \
		--image="rancher/k3s:$(K3S_VERSION)" \
		--registry-create=kxds-registry.localhost:0.0.0.0:5000 \
		kxds-dev

.PHONY: delete_cluster
delete_cluster: ## Delete the dev cluster
	k3d cluster delete kxds-dev

.PHONY: deploy_crds
deploy_crds: ## Deploy the kudo CRDs in dev cluster
	kubectl apply -f ./helm/crds

.PHONY: deploy
deploy: deploy_crds
	helm template \
		--values helm/values.yaml \
		--set image.devRef=ko://github.com/jlevesy/kxds/cmd/controller \
		kxds-dev ./helm | KO_DOCKER_REPO=kxds-registry.localhost:5000 ko apply -B -t dev -f -

##@ Build

.PHONY: build
build: generate fmt vet ## Build controller binary.
	go build -o bin/controller ./cmd/controller

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/controller

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.9.2

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
