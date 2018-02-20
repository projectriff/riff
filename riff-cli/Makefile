.PHONY: build clean
OUTPUT = riff
OUTPUT_LINUX = $(OUTPUT)-linux
OUTPUT_WINDOWS = $(OUTPUT)-windows
BUILD_FLAGS =

ifeq ($(OS),Windows_NT)
    detected_OS := Windows
else
    detected_OS := $(shell sh -c 'uname -s 2>/dev/null || echo not')
endif

ifeq ($(detected_OS),Linux)
	BUILD_FLAGS += -ldflags "-linkmode external -extldflags -static"
endif


GO_SOURCES = $(shell find cmd pkg -type f -name '*.go')

build: $(OUTPUT)

vendor: Gopkg.toml
	dep ensure

test: build
	go test -v ./...

$(OUTPUT): $(GO_SOURCES) vendor
	go build github.com/projectriff/riff-cli/cmd/riff

$(OUTPUT_LINUX): $(GO_SOURCES) vendor
	# This builds the executable from Go sources on *your* machine, targeting Linux OS
	# and linking everything statically, to minimize Docker image size
	# See e.g. https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/ for details
	CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS) -v -a -installsuffix cgo -o $(OUTPUT_LINUX) github.com/projectriff/riff-cli/cmd/riff

$(OUTPUT_WINDOWS): $(GO_SOURCES) vendor
	# This builds the executable from Go sources on *your* machine, targeting Windows
	# and linking everything statically, to minimize Docker image size
	# See e.g. https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/ for details
	CGO_ENABLED=0 GOOS=windows go build $(BUILD_FLAGS) -v -a -installsuffix cgo -o $(OUTPUT_WINDOWS) github.com/projectriff/riff-cli/cmd/riff

clean:
	rm -f $(OUTPUT)
	rm -f $(OUTPUT_LINUX)
	rm -f $(OUTPUT_WINDOWS)
