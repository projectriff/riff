.PHONY: all clean build test

GO_SOURCES = $(shell find . -type f -name '*.go')
OUTPUT = riff

all: test build

clean:
	rm riff

build: $(OUTPUT)

$(OUTPUT): $(GO_SOURCES)
	go build ./cmd/riff

test:
	go test -v ./...
