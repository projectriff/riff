.PHONY: build clean fetch-grpc grpc dockerize test
OUTPUT = function-sidecar
GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')
TAG = 0.0.4

GRPC_DIR = pkg/dispatcher/grpc

build: $(OUTPUT) vendor

$(OUTPUT): $(GO_SOURCES) vendor
	go build cmd/function-sidecar.go

vendor: glide.lock
	glide install -v --force

glide.lock: glide.yaml
	glide up -v --force

clean:
	rm -f $(OUTPUT)

test:
	go test -v ./...

grpc: $(GRPC_DIR)/function/function.pb.go

$(GRPC_DIR)/function/function.pb.go: $(GRPC_DIR)/function/function.proto
	protoc -I $(GRPC_DIR)/function $(GRPC_DIR)/function/function.proto --go_out=plugins=grpc:$(GRPC_DIR)/function

fetch-grpc:
	rm -f $(GRPC_DIR)/function/*.proto
	wget https://raw.githubusercontent.com/projectriff/function-proto/master/function.proto \
	     -P $(GRPC_DIR)/function

dockerize: $(GO_SOURCES) vendor
	docker build . -t projectriff/function-sidecar:$(TAG)
