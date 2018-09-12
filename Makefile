.PHONY: build clean test all release

OUTPUT = ./riff
GO_SOURCES = $(shell find . -type f -name '*.go' -not -regex '.*/mocks/.*' -not -regex '.*/vendor_mocks/.*')
VERSION ?= $(shell cat VERSION)
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff-index --quiet HEAD -- || echo "dirty")
LDFLAGS_VERSION = -X github.com/projectriff/riff/cmd/commands.cli_version=$(VERSION) \
				  -X github.com/projectriff/riff/cmd/commands.cli_gitsha=$(GITSHA) \
				  -X github.com/projectriff/riff/cmd/commands.cli_gitdirty=$(GITDIRTY)
GOBIN ?= $(shell go env GOPATH)/bin

all: test docs

build: $(OUTPUT)

test: build pkg/core/mocks pkg/core/vendor_mocks
	GO111MODULE=on go test ./...

pkg/core/mocks: pkg/core/client.go
	rm -fR pkg/core/mocks && \
	GO111MODULE=on go mod vendor && \
	mockery -output pkg/core/mocks -outpkg mocks -dir pkg/core -name Client && \
	rm -fR vendor/

pkg/core/vendor_mocks: go.sum
	rm -fR pkg/core/vendor_mocks && \
	GO111MODULE=on go mod vendor && \
	mockery -output pkg/core/vendor_mocks -outpkg vendor_mocks -dir vendor/k8s.io/client-go/kubernetes -name Interface && \
	mockery -output pkg/core/vendor_mocks -outpkg vendor_mocks -dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 -name CoreV1Interface && \
	mockery -output pkg/core/vendor_mocks -outpkg vendor_mocks -dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 -name NamespaceInterface && \
	mockery -output pkg/core/vendor_mocks -outpkg vendor_mocks -dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 -name ServiceAccountInterface && \
	mockery -output pkg/core/vendor_mocks -outpkg vendor_mocks -dir vendor/k8s.io/client-go/kubernetes/typed/core/v1 -name SecretInterface && \
	rm -fR vendor/

install: build
	cp $(OUTPUT) $(GOBIN)

$(OUTPUT): $(GO_SOURCES) VERSION
	GO111MODULE=on go build -o $(OUTPUT) -ldflags "$(LDFLAGS_VERSION)" main.go

release: $(GO_SOURCES) VERSION
	GO111MODULE=on GOOS=darwin   GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     main.go && tar -czf riff-darwin-amd64.tgz $(OUTPUT) && rm -f $(OUTPUT)
	GO111MODULE=on GOOS=linux    GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     main.go && tar -czf riff-linux-amd64.tgz $(OUTPUT) && rm -f $(OUTPUT)
	GO111MODULE=on GOOS=windows  GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT).exe main.go && zip -mq riff-windows-amd64.zip $(OUTPUT).exe && rm -f $(OUTPUT).exe

docs: $(OUTPUT)
	rm -fR docs && $(OUTPUT) docs

clean:
	rm -f $(OUTPUT)
	rm -f riff-darwin-amd64.tgz
	rm -f riff-linux-amd64.tgz
	rm -f riff-windows-amd64.zip
