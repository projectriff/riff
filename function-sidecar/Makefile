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

grpc: $(GRPC_DIR)/fntypes/fntypes.pb.go $(GRPC_DIR)/function/function.pb.go

$(GRPC_DIR)/fntypes/fntypes.pb.go: $(GRPC_DIR)/fntypes/fntypes.proto
	protoc -I $(GRPC_DIR)/fntypes fntypes.proto --go_out=$(GRPC_DIR)/fntypes
$(GRPC_DIR)/function/function.pb.go: $(GRPC_DIR)/function/function.proto $(GRPC_DIR)/fntypes/fntypes.pb.go
	protoc -I $(GRPC_DIR)/fntypes -I $(GRPC_DIR)/function function.proto --go_out=Mfntypes.proto=github.com/sk8sio/function-sidecar/pkg/dispatcher/grpc/fntypes,plugins=grpc:$(GRPC_DIR)/function

fetch-grpc:
	rm -f $(GRPC_DIR)/fntypes/*.proto $(GRPC_DIR)/function/*.proto
	wget https://raw.githubusercontent.com/markfisher/sk8s/master/function-proto/src/main/proto/fntypes.proto \
		-P $(GRPC_DIR)/fntypes
	wget https://raw.githubusercontent.com/markfisher/sk8s/master/function-proto/src/main/proto/function.proto \
	     -P $(GRPC_DIR)/function

dockerize: build-for-docker
	docker build . -t sk8s/function-sidecar:0.0.1-SNAPSHOT
