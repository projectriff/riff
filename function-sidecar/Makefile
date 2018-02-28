.PHONY: build clean grpc dockerize test
OUTPUT = function-sidecar
GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')
TAG ?= $(shell cat ../VERSION)

build: $(OUTPUT)

$(OUTPUT): $(GO_SOURCES) ../vendor
	go build cmd/function-sidecar.go

clean:
	rm -f $(OUTPUT)

test: ../vendor
	go test -v ./...

dockerize: $(GO_SOURCES) ../vendor
	docker build .. --build-arg COMPONENT=function-sidecar -t projectriff/function-sidecar:$(TAG)

pkg/dispatcher/grpc/function/function.pb.go: ../function-proto/function.proto
	protoc -I ../function-proto ../function-proto/function.proto --go_out=plugins=grpc:pkg/dispatcher/grpc/function

../vendor: ../glide.lock
	cd .. && glide install -v --force

../glide.lock: ../glide.yaml
	cd .. && glide up -v --force

