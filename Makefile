.PHONY: build clean test all release gen-mocks

OUTPUT = ./riff
GO_SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= $(shell cat VERSION)
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff-index --quiet HEAD -- || echo "dirty")
LDFLAGS_VERSION = -X github.com/projectriff/riff/cmd/commands.cli_version=$(VERSION) \
				  -X github.com/projectriff/riff/cmd/commands.cli_gitsha=$(GITSHA) \
				  -X github.com/projectriff/riff/cmd/commands.cli_gitdirty=$(GITDIRTY)
GOBIN ?= $(shell go env GOPATH)/bin

all: build test docs

build: $(OUTPUT)

test:
	go test ./...

gen-mocks:
	mockery -output pkg/core/mocks 			-outpkg mocks 			-dir pkg/core 											-name Client
	mockery -output pkg/core/mocks 			-outpkg mocks 			-dir pkg/core 											-name ImageClient
	mockery -output pkg/core/vendor_mocks 	-outpkg vendor_mocks 	-dir vendor/k8s.io/client-go/kubernetes 				-name Interface
	mockery -output pkg/core/vendor_mocks 	-outpkg vendor_mocks 	-dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 	-name CoreV1Interface
	mockery -output pkg/core/vendor_mocks 	-outpkg vendor_mocks 	-dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 	-name NamespaceInterface
	mockery -output pkg/core/vendor_mocks 	-outpkg vendor_mocks 	-dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 	-name ServiceAccountInterface
	mockery -output pkg/core/vendor_mocks 	-outpkg vendor_mocks 	-dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 	-name SecretInterface

install: build
	cp $(OUTPUT) $(GOBIN)

$(OUTPUT): $(GO_SOURCES) vendor VERSION
	go build -o $(OUTPUT) -ldflags "$(LDFLAGS_VERSION)" main.go

release: $(GO_SOURCES) vendor VERSION
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

