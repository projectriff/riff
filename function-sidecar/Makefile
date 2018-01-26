.PHONY: build clean fetch-grpc grpc dockerize
OUTPUT = function-sidecar
OUTPUT_LINUX = function-sidecar-linux
BUILD_FLAGS =

ifeq ($(OS),Windows_NT)
    detected_OS := Windows
else
    detected_OS := $(shell sh -c 'uname -s 2>/dev/null || echo not')
endif

ifeq ($(detected_OS),Linux)
        BUILD_FLAGS += -ldflags "-linkmode external -extldflags -static"
endif


GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')

GRPC_DIR = pkg/dispatcher/grpc

build: $(OUTPUT)

build-for-docker: $(OUTPUT_LINUX)

$(OUTPUT): $(GO_SOURCES) vendor
	go build cmd/function-sidecar.go

$(OUTPUT_LINUX): $(GO_SOURCES) vendor
	# This builds the executable from Go sources on *your* machine, targeting Linux OS
	# and linking everything statically, to minimize Docker image size
	# See e.g. https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/ for details
	CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS) -v -a -installsuffix cgo -o $(OUTPUT_LINUX) cmd/function-sidecar.go

vendor: Gopkg.toml
	dep ensure

clean:
	rm -f $(OUTPUT)
	rm -f $(OUTPUT_LINUX)

test: build
	go test -v ./...

grpc: $(GRPC_DIR)/function/function.pb.go

$(GRPC_DIR)/function/function.pb.go: $(GRPC_DIR)/function/function.proto
	protoc -I $(GRPC_DIR)/function $(GRPC_DIR)/function/function.proto --go_out=plugins=grpc:$(GRPC_DIR)/function

fetch-grpc:
	rm -f $(GRPC_DIR)/function/*.proto
	wget https://raw.githubusercontent.com/projectriff/function-proto/master/function.proto \
	     -P $(GRPC_DIR)/function

dockerize: build-for-docker
	docker build . -t projectriff/function-sidecar:0.0.3
