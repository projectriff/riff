.PHONY: all build test gen-mocks verify-mocks

GENERATED_PATH = pkg/transport/mocktransport/\*.go
GO_SOURCES = $(shell find pkg -type f -name '*.go' ! -path $(GENERATED_PATH))
GENERATED_SOURCES = $(shell find pkg -type f -name '*.go' -path $(GENERATED_PATH))
PKGS = $(shell go list ./pkg/...)

all: verify-mocks build
	@echo "To run the tests ensure kafka and zookeeper are running locally then issue 'make test'"

build: $(GO_SOURCES) ../vendor
	go build $(PKGS)

test: $(GO_SOURCES) ../vendor
	go test ./...

gen-mocks $(GENERATED_SOURCES): $(GO_SOURCES) ../vendor
	go get -u github.com/vektra/mockery/.../
	go generate ./...

# verify generated mocks which are committed or staged are up to date
verify-mocks: gen-mocks
	git diff --exit-code $(GENERATED_SOURCES)

../vendor: ../glide.lock
	cd .. && glide install -v --force

../glide.lock: ../glide.yaml
	cd .. && glide up -v --force

