.PHONY: build clean test all release gen-mocks

OUTPUT = ./riff-distro
GO_SOURCES = $(shell find .. -type f -name '*.go' -not -regex '.*/mocks/.*' -not -regex '.*/vendor_mocks/.*')
GOBIN ?= $(shell go env GOPATH)/bin
VERSION ?= $(shell cat ../VERSION)
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff-index --quiet HEAD -- || echo "dirty")
LDFLAGS_VERSION = -X github.com/projectriff/riff/distro/commands.cli_version=$(VERSION) \
				  -X github.com/projectriff/riff/distro/commands.cli_gitsha=$(GITSHA) \
				  -X github.com/projectriff/riff/distro/commands.cli_gitdirty=$(GITDIRTY)

all: build test

build: $(OUTPUT)

test: build
	go test ./...

install: build
	cp $(OUTPUT) $(GOBIN)

$(OUTPUT): $(GO_SOURCES) ../vendor ../VERSION
	go build -o $(OUTPUT) -ldflags "$(LDFLAGS_VERSION)" main.go

clean:
	rm -f $(OUTPUT)

../vendor: ../Gopkg.lock
	cd .. && dep ensure -vendor-only && touch vendor

../Gopkg.lock: ../Gopkg.toml
	cd .. && dep ensure -no-vendor && touch Gopkg.lock

