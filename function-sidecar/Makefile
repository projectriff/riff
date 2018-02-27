.PHONY: build clean grpc dockerize test
OUTPUT = function-sidecar
GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')
TAG = 0.0.5-snapshot

build: $(OUTPUT)

$(OUTPUT): $(GO_SOURCES)
	go build cmd/function-sidecar.go

clean:
	rm -f $(OUTPUT)

test:
	go test -v ./...

dockerize: $(GO_SOURCES) vendor
	docker build . -t projectriff/function-sidecar:$(TAG)

pkg/dispatcher/grpc/function/function.pb.go: ../function-proto/function.proto
	protoc -I ../function-proto ../function-proto/function.proto --go_out=plugins=grpc:pkg/dispatcher/grpc/function

