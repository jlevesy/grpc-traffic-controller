TEST_COUNT?=1
TEST_PKG?=gtc
K3S_VERSION?=v1.25.0-k3s1
CODE_GENERATOR_VERSION=0.27.3
CONTROLLER_TOOLS_VERSION=0.9.2
LOG_LEVEL?=info

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: test
test: generate gen_protoc fast_test ## Regenerate proto and run tests.

.PHONY: fast_test
fast_test:  ## Run tests.
	LOG_LEVEL=$(LOG_LEVEL) GRPC_XDS_BOOTSTRAP=$(PWD)/pkg/echoserver/xds-bootstrap.json go test ./$(TEST_PKG) -cover -count=$(TEST_COUNT) -v -run="$(T)"

.PHONY: debug_test
debug_test: generate gen_protoc ## Run tests with delve.
	LOG_LEVEL=$(LOG_LEVEL) GRPC_XDS_BOOTSTRAP=$(PWD)/pkg/echoserver/xds-bootstrap.json dlv test ./$(TEST_PKG) -- -test.count=$(TEST_COUNT) -test.v -test.run="$(T)"

.PHONY: ci_test
ci_test: ## Run tests without generation.
	GRPC_XDS_BOOTSTRAP=$(PWD)/pkg/echoserver/xds-bootstrap.json go test ./... -cover -count=$(TEST_COUNT) -v

.PHONY: dev
dev: create_cluster deploy install_example

CMD ?= ash

.PHONY: client_shell_0
client_shell_0:
	kubectl exec -n echo-client -ti $(shell kubectl -n echo-client get pods  -o jsonpath="{.items[0].metadata.name}") -- $(CMD)

.PHONY: client_shell_1
client_shell_1:
	kubectl exec -n echo-client -ti $(shell kubectl -n echo-client get pods  -o jsonpath="{.items[1].metadata.name}") -- $(CMD)

.PHONY: debug_client
debug_client:
	GRPC_GO_LOG_VERBOSITY_LEVEL=99 GRPC_GO_LOG_SEVERITY_LEVEL=info GRPC_XDS_BOOTSTRAP=./pkg/echoserver/xds-bootstrap.json dlv debug ./example/cmd/client -- -addr xds:///echo-server/basic coucou

.PHONY: local_client
local_client:
	GRPC_GO_LOG_VERBOSITY_LEVEL=99 GRPC_GO_LOG_SEVERITY_LEVEL=info GRPC_XDS_BOOTSTRAP=./pkg/echoserver/xds-bootstrap.json go run ./example/cmd/client -addr xds:///echo-server/basic coucou


.PHONY: install_example
install_example: gen_protoc ## install an example in the current cluster
	KO_DOCKER_REPO=gtc-registry.localhost:5000 ko apply -f ./example/k8s/echo-server
	KO_DOCKER_REPO=gtc-registry.localhost:5000 ko apply -f ./example/k8s/echo-client

.PHONY: create_cluster
create_cluster: ## run a local k3d cluster
	k3d cluster create \
		--image="rancher/k3s:$(K3S_VERSION)" \
		--port "16000:30000@server:0" \
		--registry-create=gtc-registry.localhost:0.0.0.0:5000 \
		gtc-dev

.PHONY: delete_cluster
delete_cluster: ## Delete the dev cluster
	k3d cluster delete gtc-dev

.PHONY: deploy_crds
deploy_crds: ## Deploy the kudo CRDs in dev cluster
	kubectl apply -f ./helm/crds


.PHONY: deploy
deploy: generate deploy_crds
	helm template \
		--values helm/values.yaml \
		--set image.devRef=ko://github.com/jlevesy/grpc-traffic-controller/cmd/controller \
		--set logLevel=$(LOG_LEVEL) \
		gtc-dev ./helm | KO_DOCKER_REPO=gtc-registry.localhost:5000 ko apply -B -t dev -f -
	kubectl apply -f example/k8s/gtc-nodeport.yaml

##@ Build

.PHONY: build
build: generate  ## Build controller binary.
	go build -o bin/controller ./cmd/controller

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: generate
generate: gen_protoc codegen gen_manifests

.PHONY: gen_protoc
gen_protoc: ## Generate protoc code for the echo server.
	protoc  \
		--go_out=. \
		--go_opt=paths=source_relative \
    --go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
    pkg/echoserver/proto/echo.proto

.PHONY: codegen
codegen: ## Run code generation for CRDs
	@bash ${GOPATH}/pkg/mod/k8s.io/code-generator@v$(CODE_GENERATOR_VERSION)/generate-groups.sh \
		all \
		github.com/jlevesy/grpc-traffic-controller/client \
		github.com/jlevesy/grpc-traffic-controller/api \
		gtc:v1alpha1 \
		--go-header-file ./hack/boilerplate.go.txt

.PHONY: gen_manifests
gen_manifests: ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(LOCALBIN)/controller-gen \
		rbac:roleName='PLACEHOLDER' \
		crd webhook \
		paths="./..." \
		output:rbac:artifacts:config=helm/templates output:crd:artifacts:config=helm/crds/
	sed -i 's/PLACEHOLDER/\{\{ include \"helm.fullname\" \. \}\}-controller/g' helm/templates/role.yaml

.PHONY: install_code_generator
install_code_generator: ## Install code generator
	go install k8s.io/code-generator/cmd/...@v$(CODE_GENERATOR_VERSION)

.PHONY: install_controller_tools
install_controller_tools: $(LOCALBIN) ## Install controller-tools
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_TOOLS_VERSION)
