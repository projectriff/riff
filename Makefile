.PHONY: build clean test all
OUTPUT = ./riff
GO_SOURCES = $(shell find cmd pkg -type f -name '*.go' -not -name 'mock_*.go')
VERSION ?= $(shell cat VERSION)

all: test docs

build: $(OUTPUT)

test: build
	go test ./...

$(OUTPUT): $(GO_SOURCES) vendor VERSION
	go build -o $(OUTPUT)  -ldflags "-X github.com/projectriff/riff-cli/cmd/commands.cli_version=$(VERSION)" cmd/main.go

docs: $(OUTPUT)
	rm -fR docs && $(OUTPUT) docs

clean:
	rm -f $(OUTPUT)

vendor: Gopkg.lock
	dep ensure -vendor-only && touch vendor

Gopkg.lock: Gopkg.toml
	dep ensure -no-vendor && touch Gopkg.lock

