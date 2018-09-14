.PHONY: build clean test all release gen-mocks

OUTPUT = ./riff-distro
GO_SOURCES = $(shell find .. -type f -name '*.go' -not -regex '.*/mocks/.*' -not -regex '.*/vendor_mocks/.*')
GOBIN ?= $(shell go env GOPATH)/bin

all: build test

build: $(OUTPUT)

test: build
	go test ./...

install: build
	cp $(OUTPUT) $(GOBIN)

$(OUTPUT): $(GO_SOURCES) ../vendor
	go build -o $(OUTPUT) main.go

clean:
	rm -f $(OUTPUT)

../vendor: ../Gopkg.lock
	cd .. && dep ensure -vendor-only && touch vendor

../Gopkg.lock: ../Gopkg.toml
	cd .. && dep ensure -no-vendor && touch Gopkg.lock

