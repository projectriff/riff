.PHONY: all build test gen-mocks verify-mocks

GENERATED_PATH = pkg/transport/mocktransport/\*.go
GO_SOURCES = $(shell find pkg -type f -name '*.go' ! -path $(GENERATED_PATH))
GENERATED_SOURCES = $(shell find pkg -type f -name '*.go' -path $(GENERATED_PATH))
PKGS = $(shell go list ./pkg/...)

all: vendor verify-mocks build
	@echo "To run the tests ensure kafka and zookeeper are running locally then issue 'make test'"

vendor: glide.lock
	glide install -v --force

glide.lock: glide.yaml
	glide up -v --force

build: $(GO_SOURCES)
	go build $(PKGS)

test: $(GO_SOURCES)
	-kafka-topics --delete --zookeeper localhost:2181 --topic test-topic
	-kafka-topics --create --zookeeper localhost:2181 --topic test-topic --partitions 1 --replication-factor 1
	KAFKA_BROKERS=localhost:9092 go test -v ./...

gen-mocks $(GENERATED_SOURCE): $(GO_SOURCES)
	go get -u github.com/vektra/mockery/.../
	go generate ./...

# verify generated mocks which are committed or staged are up to date
verify-mocks: gen-mocks
	git diff --exit-code $(GENERATED_SOURCES)
