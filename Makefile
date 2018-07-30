.PHONY: build clean test all release
OUTPUT = ./riff
GO_SOURCES = $(shell find cmd pkg -type f -name '*.go' -not -name 'mock_*.go')
VERSION ?= $(shell cat VERSION)
LDFLAGS_VERSION = -X github.com/projectriff/riff/cmd/commands.cli_version=$(VERSION)
GOBIN ?= $(shell go env GOPATH)/bin

all: test docs

build: $(OUTPUT)

test: build
	go test ./...

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

