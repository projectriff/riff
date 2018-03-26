.PHONY: build clean grpc dockerize test vendor
OUTPUT = function-sidecar
GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')
TAG ?= $(shell cat ../VERSION)

build: $(OUTPUT)

$(OUTPUT): $(GO_SOURCES) vendor
	go build cmd/function-sidecar.go

clean:
	rm -f $(OUTPUT)

test: vendor
	go test ./...

dockerize: $(GO_SOURCES) vendor
	docker build .. --build-arg COMPONENT=function-sidecar -t projectriff/function-sidecar:$(TAG)

debug-dockerize: $(GO_SOURCES) vendor
	docker build .. --build-arg COMPONENT=function-sidecar -t projectriff/function-sidecar:$(TAG) -f ../Dockerfile-debug

pkg/dispatcher/grpc/function/function.pb.go: ../function-proto/function.proto
	protoc -I ../function-proto ../function-proto/function.proto --go_out=plugins=grpc:pkg/dispatcher/grpc/function

vendor:
	$(MAKE) -C .. vendor
