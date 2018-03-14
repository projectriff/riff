.PHONY: build clean test release verify-docs all
OUTPUT = riff
OUTPUT_WINDOWS = $(OUTPUT).exe
GO_SOURCES = $(shell find cmd pkg -type f -name '*.go')
VERSION ?= $(shell cat ../VERSION)


all: build docs test

build: $(OUTPUT)

test: build
	go test ./...

$(OUTPUT): $(GO_SOURCES) ../vendor
	go build -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT) ../cli.go

release: $(GO_SOURCES) ../vendor
	GOOS=darwin   GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT) ../cli.go && tar -czf riff-darwin-amd64.tgz $(OUTPUT) && rm -f $(OUTPUT)
	GOOS=linux    GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT) ../cli.go && tar -czf riff-linux-amd64.tgz $(OUTPUT) && rm -f $(OUTPUT)
	GOOS=windows  GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT_WINDOWS) ../cli.go && zip -mq riff-windows-amd64.zip $(OUTPUT_WINDOWS)

docs: $(OUTPUT)
	rm -fR docs && ./$(OUTPUT) docs

verify-docs: docs
	git diff --exit-code docs

clean:
	rm -f $(OUTPUT)
	rm -f riff-darwin-amd64.tgz
	rm -f riff-linux-amd64.tgz
	rm -f riff-windows-amd64.zip

../vendor: ../glide.lock
	cd .. && glide install -v --force

../glide.lock: ../glide.yaml
	cd .. && glide up -v --force

