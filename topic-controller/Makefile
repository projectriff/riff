.PHONY: build build-for-docker clean dockerize
OUTPUT = topic-controller
OUTPUT_LINUX = $(OUTPUT)-linux

GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')

build: $(OUTPUT)

build-for-docker: $(OUTPUT_LINUX)

test: build
	go test -v ./...

$(OUTPUT): $(GO_SOURCES) vendor
	go build cmd/topic-controller.go

$(OUTPUT_LINUX): $(GO_SOURCES) vendor
	# This builds the executable from Go sources on *your* machine, targeting Linux OS
	# and linking everything statically, to minimize Docker image size
	# See e.g. https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/ for details
	CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o $(OUTPUT_LINUX) cmd/topic-controller.go

vendor: Gopkg.toml
	dep ensure

clean:
	rm -f $(OUTPUT)
	rm -f $(OUTPUT_LINUX)

dockerize: build-for-docker
	docker build . -t sk8s/topic-controller:0.0.1-SNAPSHOT
