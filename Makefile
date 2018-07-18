.PHONY: build clean test
OUTPUT = ./riff
GO_SOURCES = $(shell find cmd pkg -type f -name '*.go' -not -name 'mock_*.go')

build: $(OUTPUT)

test: build
	go test ./...

$(OUTPUT): $(GO_SOURCES) vendor
	go build -o $(OUTPUT) cmd/main.go

docs: $(OUTPUT)
	rm -fR docs && $(OUTPUT) docs

clean:
	rm -f $(OUTPUT)

vendor: Gopkg.lock
	dep ensure -vendor-only && touch vendor

Gopkg.lock: Gopkg.toml
	dep ensure -no-vendor && touch Gopkg.lock

