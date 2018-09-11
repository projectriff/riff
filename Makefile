.PHONY: build clean test all release gen-mocks

OUTPUT = ./riff
GO_SOURCES = $(shell find cmd pkg -type f -name '*.go' -not -regex '.*/mocks/.*' -not -regex '.*/vendor_mocks/.*')
VERSION ?= $(shell cat VERSION)
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff-index --quiet HEAD -- || echo "dirty")
LDFLAGS_VERSION = -X github.com/projectriff/riff/cmd/commands.cli_version=$(VERSION) \
				  -X github.com/projectriff/riff/cmd/commands.cli_gitsha=$(GITSHA) \
				  -X github.com/projectriff/riff/cmd/commands.cli_gitdirty=$(GITDIRTY)
GOBIN ?= $(shell go env GOPATH)/bin

all: test docs

build: $(OUTPUT)

test: build gen-mocks
	go test ./...

pkg/core/mocks/Client.go: pkg/core/client.go
	mockery -output pkg/core/mocks -outpkg mocks -dir pkg/core -name Client

pkg/core/vendor_mocks/Interface.go: $(shell find vendor/k8s.io/client-go/kubernetes -type f)
	mockery -output pkg/core/vendor_mocks -outpkg vendor_mocks -dir vendor/k8s.io/client-go/kubernetes -name Interface

pkg/core/vendor_mocks/CoreV1Interface.go \
pkg/core/vendor_mocks/NamespaceInterface.go \
pkg/core/vendor_mocks/ServiceAccountInterface.go \
pkg/core/vendor_mocks/SecretInterface.go : $(shell find vendor/k8s.io/client-go/kubernetes/typed/core/v1 -type f)
	mockery -output pkg/core/vendor_mocks -outpkg vendor_mocks -dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 -name $(notdir $(basename $@))

gen-mocks: pkg/core/mocks/Client.go $(wildcard pkg/core/vendor_mocks/*.go)

install: build
	cp $(OUTPUT) $(GOBIN)

$(OUTPUT): $(GO_SOURCES) main.go vendor VERSION
	go build -o $(OUTPUT) -ldflags "$(LDFLAGS_VERSION)" main.go

release: $(GO_SOURCES) main.go vendor VERSION
	GOOS=darwin   GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     main.go && tar -czf riff-darwin-amd64.tgz $(OUTPUT) && rm -f $(OUTPUT)
	GOOS=linux    GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     main.go && tar -czf riff-linux-amd64.tgz $(OUTPUT) && rm -f $(OUTPUT)
	GOOS=windows  GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT).exe main.go && zip -mq riff-windows-amd64.zip $(OUTPUT).exe && rm -f $(OUTPUT).exe

docs: $(OUTPUT)
	rm -fR docs && $(OUTPUT) docs

clean:
	rm -f $(OUTPUT)
	rm -f riff-darwin-amd64.tgz
	rm -f riff-linux-amd64.tgz
	rm -f riff-windows-amd64.zip

vendor: Gopkg.lock
	dep ensure -vendor-only && touch vendor

Gopkg.lock: Gopkg.toml
	dep ensure -no-vendor && touch Gopkg.lock

