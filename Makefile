OUTPUT = ./riff
GO_SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= $(shell cat VERSION)
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff --quiet HEAD || echo "dirty")
LDFLAGS_VERSION = -X github.com/projectriff/riff/pkg/cli.cli_name=riff \
				  -X github.com/projectriff/riff/pkg/cli.cli_version=$(VERSION) \
				  -X github.com/projectriff/riff/pkg/cli.cli_gitsha=$(GITSHA) \
				  -X github.com/projectriff/riff/pkg/cli.cli_gitdirty=$(GITDIRTY)

.PHONY: all
all: build test docs

.PHONY: clean
clean:
	rm -f $(OUTPUT)
	rm -f riff-darwin-amd64.tgz
	rm -f riff-linux-amd64.tgz
	rm -f riff-windows-amd64.zip

.PHONY: build
build: $(OUTPUT)

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -v --race -coverprofile=coverage.txt -covermode=atomic ./...

$(OUTPUT): $(GO_SOURCES) VERSION
	go build -o $(OUTPUT) -ldflags "$(LDFLAGS_VERSION)" ./cmd/riff

.PHONY: release
release: $(GO_SOURCES) VERSION
	GOOS=darwin   GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     ./cmd/riff && tar -czf riff-darwin-amd64.tgz $(OUTPUT)     && rm -f $(OUTPUT)
	GOOS=linux    GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT)     ./cmd/riff && tar -czf riff-linux-amd64.tgz  $(OUTPUT)     && rm -f $(OUTPUT)
	GOOS=windows  GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(OUTPUT).exe ./cmd/riff && zip -mq riff-windows-amd64.zip $(OUTPUT).exe && rm -f $(OUTPUT).exe

docs: $(OUTPUT) clean-docs
	$(OUTPUT) docs

.PHONY: verify-docs
verify-docs: docs
	git diff --exit-code docs

.PHONY: clean-docs
clean-docs:
	rm -fR docs

.PHONY: check-mockery
check-mockery:
    # Use go get in GOPATH mode to install/update mockery. This avoids polluting go.mod/go.sum.
	@which mockery > /dev/null || (echo mockery not found: issue \"GO111MODULE=off go get -u  github.com/vektra/mockery/.../\" && false)

.PHONY: gen-mocks
gen-mocks: check-mockery
	mockery -output ./pkg/testing/pack -outpkg pack -dir ./pkg/pack -name Client
	mockery -output ./pkg/testing/kail -outpkg kail -dir ./pkg/kail -name Logger
