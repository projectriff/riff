.PHONY: clean gen-mocks build test help

OUTPUT = ./provisioner
GO_SOURCES = $(shell find . -type f -name '*.go')
GOBIN ?= $(shell go env GOPATH)/bin

.DEFAULT_GOAL := help

clean: ## remove the binary
	rm -f $(OUTPUT)

gen-mocks: ## generate mocks
	go generate ./...

build: gen-mocks $(OUTPUT) ## build the project binary

test: ## run the project tests
	go test -v ./...

$(OUTPUT): $(GO_SOURCES)
	go build -v -o $(OUTPUT) cmd/provisioner/main.go

# source: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'