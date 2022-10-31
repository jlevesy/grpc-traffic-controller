TEST_COUNT?=1
K3S_VERSION=v1.25.0-k3s1
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
SHELL=/usr/bin/env bash -o pipefail
.SHELLFLAGS=-ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName='PLACEHOLDER' crd webhook paths="./..." output:rbac:artifacts:config=helm/templates output:crd:artifacts:config=helm/crds/
	sed -i 's/PLACEHOLDER/\{\{ include \"helm.fullname\" \. \}\}-controller/g' helm/templates/role.yaml

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: gen-protoc
gen-protoc: ## Generate protoc code for the echo server.
	protoc  \
		--go_out=. \
		--go_opt=paths=source_relative \
    --go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
    pkg/echoserver/proto/echo.proto

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate gen-protoc fmt vet ## Run tests.
	GRPC_XDS_BOOTSTRAP=$(PWD)/pkg/echoserver/xds-bootstrap.json go test ./... -cover -count=$(TEST_COUNT) -v

.PHONY: ci_test
ci_test: ## Run tests without generation.
	GRPC_XDS_BOOTSTRAP=$(PWD)/pkg/echoserver/xds-bootstrap.json go test ./... -cover -count=$(TEST_COUNT) -v

.PHONY: dev
dev: create_cluster deploy install_example

.PHONY: client_shell
client_shell:
	kubectl exec -n echo-client -ti $(shell kubectl -n echo-client get pods  -o jsonpath="{.items[0].metadata.name}") -- ash

.PHONY: install_example
install_example: gen-protoc ## install an example in the current cluster
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

.PHONY: run_local
run_local:
	GRPC_XDS_BOOTSTRAP=./pkg/echoserver/xds-bootstrap.json go run ./example/cmd/client --addr xds:///echo-server coucou

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.9.2

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
