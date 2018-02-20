.PHONY: build clean test release
OUTPUT = riff
GO_SOURCES = $(shell find cmd pkg -type f -name '*.go')

build: $(OUTPUT)

vendor: Gopkg.lock
	dep ensure -vendor-only

Gopkg.lock: Gopkg.toml
	dep ensure -update

test: build
	go test -v ./...

$(OUTPUT): $(GO_SOURCES) vendor
	go build -o $(OUTPUT) cmd/riff/main.go

release: $(GO_SOURCES) vendor
	GOOS=darwin   GOARCH=amd64 go build -o $(OUTPUT) cmd/riff/main.go && tar -czf riff-darwin-amd64.tgz  $(OUTPUT)
	GOOS=linux    GOARCH=amd64 go build -o $(OUTPUT) cmd/riff/main.go && tar -czf riff-linux-amd64.tgz   $(OUTPUT)
	GOOS=windows  GOARCH=amd64 go build -o $(OUTPUT) cmd/riff/main.go && zip -mq riff-windows-amd64.zip $(OUTPUT)

$(OUTPUT_LINUX): $(GO_SOURCES) vendor
	GOOS=linux go build -o $(OUTPUT_LINUX) cmd/riff/main.go

$(OUTPUT_WINDOWS): $(GO_SOURCES) vendor
	GOOS=windows go build -o $(OUTPUT_WINDOWS) cmd/riff/main.go

clean:
	rm -f $(OUTPUT)
	rm -f riff-darwin-amd64.tgz
	rm -f riff-linux-amd64.tgz
	rm -f riff-windows-amd64.zip
