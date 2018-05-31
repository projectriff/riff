.PHONY: build clean grpc dockerize test
OUTPUT = function-sidecar
GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')
TAG ?= $(shell cat ../VERSION)

build: $(OUTPUT)

$(OUTPUT): $(GO_SOURCES) ../vendor
	go build cmd/function-sidecar.go

clean:
	rm -f $(OUTPUT) .dockerize

test: ../vendor
	go test ./...

dockerize: .dockerize

.dockerize: $(GO_SOURCES) ../vendor ../Dockerfile ../.dockerignore
	docker build .. --build-arg COMPONENT=function-sidecar -t projectriff/function-sidecar:$(TAG) && touch .dockerize

debug-dockerize: $(GO_SOURCES) ../vendor
	docker build .. --build-arg COMPONENT=function-sidecar -t projectriff/function-sidecar:$(TAG) -f ../Dockerfile-debug

pkg/dispatcher/grpc/function/function.pb.go: ../function-proto/function.proto
	protoc -I ../function-proto ../function-proto/function.proto --go_out=plugins=grpc:pkg/dispatcher/grpc/function

../vendor: ../glide.lock
	cd .. && glide install -v --force

../glide.lock: ../glide.yaml
	cd .. && glide up -v --force

