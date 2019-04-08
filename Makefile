.PHONY: build clean test all release gen-mocks check-mockery verify-docs clean-docs

OUTPUT = ./riff
GO_SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= $(shell cat VERSION)
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff --quiet HEAD || echo "dirty")
LDFLAGS_VERSION = -X github.com/projectriff/riff/pkg/env.cli_name=riff \
				  -X github.com/projectriff/riff/pkg/env.cli_version=$(VERSION) \
				  -X github.com/projectriff/riff/pkg/env.cli_gitsha=$(GITSHA) \
				  -X github.com/projectriff/riff/pkg/env.cli_gitdirty=$(GITDIRTY)
GOBIN ?= $(shell go env GOPATH)/bin

all: build test docs

build: $(OUTPUT)

test:
	GO111MODULE=on go test ./...

check-mockery:
	@which mockery > /dev/null || (echo mockery not found: issue \"GO111MODULE=off go get -u  github.com/vektra/mockery/.../\" && false)

check-jq:
	@which jq > /dev/null || (echo jq not found: please install jq, eg \"brew install jq\" && false)

gen-mocks: check-mockery check-jq
	GO111MODULE=on mockery -output pkg/core/mocks/mockbuilder        -outpkg mockbuilder   -dir pkg/core                                                                                           -name Builder
	GO111MODULE=on mockery -output pkg/core/mocks                    -outpkg mocks         -dir pkg/core                                                                                           -name Client
	GO111MODULE=on mockery -output pkg/core/kustomize/mocks          -outpkg mockkustomize -dir pkg/core/kustomize                                                                                 -name Kustomizer
	GO111MODULE=on mockery -output pkg/core/vendor_mocks/mockbuild   -outpkg mockbuild     -dir $(call source_of,github.com/knative/build)/pkg/client/clientset/versioned                          -name Interface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks/mockbuild   -outpkg mockbuild     -dir $(call source_of,github.com/knative/build)/pkg/client/clientset/versioned/typed/build/v1alpha1     -name BuildV1alpha1Interface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks/mockbuild   -outpkg mockbuild     -dir $(call source_of,github.com/knative/build)/pkg/client/clientset/versioned/typed/build/v1alpha1     -name ClusterBuildTemplateInterface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks/mockserving -outpkg mockserving   -dir $(call source_of,github.com/knative/serving)/pkg/client/clientset/versioned                        -name Interface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks/mockserving -outpkg mockserving   -dir $(call source_of,github.com/knative/serving)/pkg/client/clientset/versioned/typed/serving/v1alpha1 -name ServingV1alpha1Interface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks/mockserving -outpkg mockserving   -dir $(call source_of,github.com/knative/serving)/pkg/client/clientset/versioned/typed/serving/v1alpha1 -name ServiceInterface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks/mockserving -outpkg mockserving   -dir $(call source_of,github.com/knative/serving)/pkg/client/clientset/versioned/typed/serving/v1alpha1 -name RevisionInterface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks             -outpkg vendor_mocks  -dir $(call source_of,k8s.io/client-go)/kubernetes/typed/core/v1                                        -name CoreV1Interface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks             -outpkg vendor_mocks  -dir $(call source_of,k8s.io/client-go)/kubernetes/typed/core/v1                                        -name NamespaceInterface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks             -outpkg vendor_mocks  -dir $(call source_of,k8s.io/client-go)/kubernetes/typed/core/v1                                        -name ServiceAccountInterface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks             -outpkg vendor_mocks  -dir $(call source_of,k8s.io/client-go)/kubernetes/typed/core/v1                                        -name ConfigMapInterface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks             -outpkg vendor_mocks  -dir $(call source_of,k8s.io/client-go)/kubernetes/typed/core/v1                                        -name SecretInterface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks             -outpkg vendor_mocks  -dir $(call source_of,k8s.io/client-go)/kubernetes                                                      -name Interface
	GO111MODULE=on mockery -output pkg/core/vendor_mocks             -outpkg vendor_mocks  -dir $(call source_of,k8s.io/client-go)/tools/clientcmd                                                 -name ClientConfig
	GO111MODULE=on mockery -output pkg/kubectl/mocks                 -outpkg mockkubectl   -dir pkg/kubectl                                                                                        -name KubeCtl

install: build
	cp $(OUTPUT) $(GOBIN)

$(OUTPUT): $(GO_SOURCES) VERSION
	GO111MODULE=on go build -o $(OUTPUT) -ldflags "$(LDFLAGS_VERSION)" ./cmd/riff

release: $(GO_SOURCES) VERSION
	GOOS=darwin   GOARCH=amd64 GO111MODULE=on go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     ./cmd/riff && tar -czf riff-darwin-amd64.tgz $(OUTPUT)     && rm -f $(OUTPUT)
	GOOS=linux    GOARCH=amd64 GO111MODULE=on go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     ./cmd/riff && tar -czf riff-linux-amd64.tgz  $(OUTPUT)     && rm -f $(OUTPUT)
	GOOS=windows  GOARCH=amd64 GO111MODULE=on go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT).exe ./cmd/riff && zip -mq riff-windows-amd64.zip $(OUTPUT).exe && rm -f $(OUTPUT).exe

docs: $(OUTPUT) clean-docs
	$(OUTPUT) docs

verify-docs: docs
	git diff --exit-code docs

clean-docs:
	rm -fR docs

clean:
	rm -f $(OUTPUT)
	rm -f riff-darwin-amd64.tgz
	rm -f riff-linux-amd64.tgz
	rm -f riff-windows-amd64.zip

define source_of
	$(shell GO111MODULE=on go mod download -json | jq -r 'select(.Path == "$(1)").Dir' | tr '\\' '/'  2> /dev/null)
endef
