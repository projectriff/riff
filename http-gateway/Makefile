.PHONY: build clean fetch-grpc grpc dockerize
OUTPUT = http-gateway

test: build
	go test -v ./...

build: $(OUTPUT)

$(OUTPUT): vendor
	go build cmd/http-gateway.go

vendor: Gopkg.toml
	dep ensure

clean:
	rm -f $(OUTPUT)

dockerize:
	# This builds the executable from Go sources on *your* machine, targeting Linux OS
	# and linking everything statically, to minimize Docker image size
	# See e.g. https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/ for details
	CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o $(OUTPUT) cmd/http-gateway.go
	docker build . -t sk8s/http-gateway:0.0.1-SNAPSHOT -t sk8s/topic-gateway:0.0.1-SNAPSHOT
