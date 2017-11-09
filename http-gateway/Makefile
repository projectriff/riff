.PHONY: build build-for-docker clean dockerize
OUTPUT = http-gateway
OUTPUT_LINUX = http-gateway-linux

GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')

build: $(OUTPUT)

build-for-docker: $(OUTPUT_LINUX)

test: build
	go test -v ./...

.arch-linux: vendor

$(OUTPUT): $(GO_SOURCES) vendor
	go build cmd/http-gateway.go

$(OUTPUT_LINUX): $(GO_SOURCES) vendor
	# This builds the executable from Go sources on *your* machine, targeting Linux OS
	# and linking everything statically, to minimize Docker image size
	# See e.g. https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/ for details
	CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o $(OUTPUT_LINUX) cmd/http-gateway.go

vendor: Gopkg.toml
	dep ensure

clean:
	rm -f $(OUTPUT)
	rm -f $(OUTPUT_LINUX)

dockerize: build-for-docker
	docker build . -t sk8s/http-gateway:0.0.1-SNAPSHOT -t sk8s/topic-gateway:0.0.1-SNAPSHOT
