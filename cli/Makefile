OUTPUT = ./riff
GO_SOURCES = $(shell find . -type f -name '*.go')
GOBIN ?= $(shell go env GOPATH)/bin
VERSION ?= $(shell cat VERSION)
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff --quiet HEAD || echo "dirty")
LDFLAGS_VERSION = -X github.com/projectriff/cli/pkg/cli.cli_name=riff \
				  -X github.com/projectriff/cli/pkg/cli.cli_version=$(VERSION) \
				  -X github.com/projectriff/cli/pkg/cli.cli_gitsha=$(GITSHA) \
				  -X github.com/projectriff/cli/pkg/cli.cli_gitdirty=$(GITDIRTY) \
				  -X github.com/projectriff/cli/pkg/cli.cli_runtimes=core,knative,streaming

.PHONY: all
all: build test verify-goimports docs ## Build, test, verify source formatting and regenerate docs

.PHONY: clean
clean: ## Delete build output
	rm -f $(OUTPUT)
	rm -f riff-darwin-amd64.tgz
	rm -f riff-linux-amd64.tgz
	rm -f riff-windows-amd64.zip

.PHONY: build
build: $(OUTPUT) ## Build riff

.PHONY: test
test: ## Run the tests
	go test ./...

.PHONY: install
install: build ## Copy build to GOPATH/bin
	cp $(OUTPUT) $(GOBIN)

.PHONY: coverage
coverage: ## Run the tests with coverage and race detection
	go test -v --race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: check-goimports
check-goimports: ## Checks if goimports is installed
	@which goimports > /dev/null || (echo goimports not found: issue \"GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports\" && false)

.PHONY: goimports
goimports: check-goimports ## Runs goimports on the project
	@goimports -w pkg cmd

.PHONY: verify-goimports
verify-goimports: check-goimports ## Verifies if all source files are formatted correctly
	@goimports -l pkg cmd | (! grep .) || (echo above files are not formatted correctly. please run \"make goimports\" && false)

$(OUTPUT): $(GO_SOURCES) VERSION
	go build -o $(OUTPUT) -ldflags "$(LDFLAGS_VERSION)" ./cmd/riff

.PHONY: release
release: $(GO_SOURCES) VERSION ## Cross-compile riff for various operating systems
	GOOS=darwin   GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     ./cmd/riff && tar -czf riff-darwin-amd64.tgz  $(OUTPUT)     && rm -f $(OUTPUT)
	GOOS=linux    GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     ./cmd/riff && tar -czf riff-linux-amd64.tgz   $(OUTPUT)     && rm -f $(OUTPUT)
	GOOS=windows  GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT).exe ./cmd/riff && zip -mq  riff-windows-amd64.zip $(OUTPUT).exe && rm -f $(OUTPUT).exe

docs: $(OUTPUT) clean-docs ## Generate documentation
	$(OUTPUT) docs

.PHONY: verify-docs
verify-docs: docs ## Verify the generated docs are up to date
	git diff --exit-code docs

.PHONY: clean-docs
clean-docs: ## Delete the generated docs
	rm -fR docs

.PHONY: check-mockery
check-mockery:
    # Use go get in GOPATH mode to install/update mockery. This avoids polluting go.mod/go.sum.
	@which mockery || (echo mockery not found: issue \"GO111MODULE=off go get -u  github.com/vektra/mockery/.../\" && false)

.PHONY: gen-mocks
gen-mocks: check-mockery clean-mocks ## Generate mocks
	mockery -output ./pkg/testing/pack -outpkg pack -dir ./pkg/pack -name Client
	mockery -output ./pkg/testing/kail -outpkg kail -dir ./pkg/kail -name Logger
	make goimports

.PHONY: clean-mocks
clean-mocks: ## Delete mocks
	rm -fR pkg/testing/pack
	rm -fR pkg/testing/kail

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
