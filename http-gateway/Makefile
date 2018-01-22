.PHONY: build clean dockerize test
OUTPUT = http-gateway

GO_SOURCES = $(shell find cmd -type f -name '*.go')
TAG = 0.0.4-snapshot

build: $(OUTPUT)

test: build
	go test -v ./...

$(OUTPUT): $(GO_SOURCES) vendor
	go build cmd/http-gateway.go

vendor: Gopkg.toml
	dep ensure

clean:
	rm -f $(OUTPUT)

dockerize: $(GO_SOURCES) vendor
	docker build . -t projectriff/http-gateway:$(TAG)

debug-dockerize: $(GO_SOURCES) vendor
	# Need to remove probes as delve starts app in paused state
	-kubectl patch deploy/http-gateway --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/livenessProbe"}]'
	-kubectl patch deploy/http-gateway --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/readinessProbe"}]'
	docker build . -t projectriff/http-gateway:$(TAG) -f Dockerfile-debug
