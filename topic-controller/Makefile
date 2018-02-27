.PHONY: build clean test dockerize debug-dockerize
OUTPUT = topic-controller
TAG = 0.0.5-snapshot

GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')

build: $(OUTPUT)

test:
	go test -v ./...

$(OUTPUT): $(GO_SOURCES)
	go build cmd/topic-controller.go

clean:
	rm -f $(OUTPUT)

dockerize: $(GO_SOURCES)
	docker build .. --build-arg COMPONENT=topic-controller -t projectriff/topic-controller:$(TAG)

debug-dockerize: $(GO_SOURCES)
	# Need to remove probes as delve starts app in paused state
	-kubectl patch deploy/topic-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/livenessProbe"}]'
	-kubectl patch deploy/topic-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/readinessProbe"}]'
	docker build . -t projectriff/topic-controller:$(TAG) -f Dockerfile-debug
