
# Target component to build/run
COMPONENT ?= build
VERSION ?= $(shell cat ../VERSION)
export VERSION_SLUG ?= $(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= crd:trivialVersions=true

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all: prepare test

.PHONY: run
run: manifests ## Run component against the configured Kubernetes cluster in ~/.kube/config
	go run ./cmd/managers/$(COMPONENT)/main.go

.PHONY: test
test: generate fmt vet manifests ## Run tests
	go test ./... -coverprofile cover.out

.PHONY: compile
compile: prepare ko ## Compile target binaries
	$(KO) resolve -L -f config/ > /dev/null

.PHONY: prepare
prepare: generate fmt vet manifests templates ## Create all generated and scaffolded files
	kustomize build config/build/default > config/riff-build.yaml
	kustomize build config/core/default > config/riff-core.yaml
	kustomize build config/knative/default > config/riff-knative.yaml
	kustomize build config/streaming/default > config/riff-streaming.yaml

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests:
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook crd:maxDescLen=0 \
		paths="./pkg/apis/build/...;./pkg/controllers/build/..." \
		output:crd:dir=./config/build/crd/bases \
		output:rbac:dir=./config/build/rbac \
		output:webhook:dir=./config/build/webhook
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook crd:maxDescLen=0 \
		paths="./pkg/apis/core/...;./pkg/controllers/core/..." \
		output:crd:dir=./config/core/crd/bases \
		output:rbac:dir=./config/core/rbac \
		output:webhook:dir=./config/core/webhook
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook crd:maxDescLen=0 \
		paths="./pkg/apis/knative/...;./pkg/controllers/knative/..." \
		output:crd:dir=./config/knative/crd/bases \
		output:rbac:dir=./config/knative/rbac \
		output:webhook:dir=./config/knative/webhook
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook crd:maxDescLen=0 \
		paths="./pkg/apis/streaming/...;./pkg/controllers/streaming/..." \
		output:crd:dir=./config/streaming/crd/bases \
		output:rbac:dir=./config/streaming/rbac \
		output:webhook:dir=./config/streaming/webhook
	# cleanup duplicate resource generation
	@rm -f config/build.*
	@rm -f config/core.*
	@rm -f config/knative.*
	@rm -f config/streaming.*

# Run go fmt against code
.PHONY: fmt
fmt: goimports
	$(GOIMPORTS) --local github.com/projectriff/riff/system -w pkg/ cmd/

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

.PHONY: generate
generate: generate-internal fmt ## Generate code

.PHONY: generate-internal
generate-internal: controller-gen mockery
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."
	$(MOCKERY) -dir ./pkg/controllers/streaming -inpkg -name StreamProvisionerClient -case snake

# find or download controller-gen, download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# find or download goimports, download goimports if necessary
goimports:
ifeq (, $(shell which goimports))
	# avoid go.* mutations from go get
	( cd / && GO111MODULE=on go get golang.org/x/tools/cmd/goimports@release-branch.go1.13 )
GOIMPORTS=$(GOBIN)/goimports
else
GOIMPORTS=$(shell which goimports)
endif

# find or download ko, download ko if necessary
ko:
ifeq (, $(shell which ko))
	GO111MODULE=off go get github.com/google/ko/cmd/ko
KO=$(GOBIN)/ko
else
KO=$(shell which ko)
endif

# find or download mockery, download mockery if necessary
mockery:
ifeq (, $(shell which mockery))
	GO111MODULE=off go get -u  github.com/vektra/mockery/.../
MOCKERY=$(GOBIN)/mockery
else
MOCKERY=$(shell which mockery)
endif

.PHONY: templates
templates: ## update templated components
	./hack/apply-template.sh config/streaming/config/bases/processor.yaml.tpl > config/streaming/config/bases/processor.yaml

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
