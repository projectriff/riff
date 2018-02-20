.PHONY: build clean test dockerize debug-dockerize
OUTPUT = topic-controller
TAG = 0.0.4

GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')

build: $(OUTPUT)

test: vendor
	go test -v ./...

$(OUTPUT): $(GO_SOURCES) vendor
	go build cmd/topic-controller.go

vendor: glide.lock
	glide install -v --force

glide.lock: glide.yaml
	glide up -v --force

clean:
	rm -f $(OUTPUT)

dockerize: $(GO_SOURCES) vendor
	docker build . -t projectriff/topic-controller:$(TAG)

debug-dockerize: $(GO_SOURCES) vendor
	# Need to remove probes as delve starts app in paused state
	-kubectl patch deploy/topic-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/livenessProbe"}]'
	-kubectl patch deploy/topic-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/readinessProbe"}]'
	docker build . -t projectriff/topic-controller:$(TAG) -f Dockerfile-debug
