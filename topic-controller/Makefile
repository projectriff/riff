.PHONY: build clean dockerize
OUTPUT = topic-controller
TAG = 0.0.4-snapshot

GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')

build: $(OUTPUT)

test:
	go test -v ./...

$(OUTPUT): $(GO_SOURCES)
	go build cmd/topic-controller.go

vendor: glide.lock
	glide install -v --force

glide.lock: glide.yaml
	glide up -v --force

clean:
	rm -f $(OUTPUT)

dockerize:
	docker build . -t projectriff/topic-controller:$(TAG)
