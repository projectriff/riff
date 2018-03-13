.PHONY: build clean test dockerize debug-dockerize kubectl-apply
OUTPUT = topic-controller
TAG ?= $(shell cat ../VERSION)

GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')

build: $(OUTPUT)

test: ../vendor
	go test ./...

$(OUTPUT): $(GO_SOURCES) ../vendor
	go build cmd/topic-controller.go

clean:
	rm -f $(OUTPUT)

dockerize: $(GO_SOURCES) ../vendor
	docker build .. --build-arg COMPONENT=topic-controller -t projectriff/topic-controller:$(TAG)

debug-dockerize: $(GO_SOURCES) ../vendor
	# Need to remove probes as delve starts app in paused state
	-kubectl patch deploy/topic-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/livenessProbe"}]'
	-kubectl patch deploy/topic-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/readinessProbe"}]'
	docker build .. --build-arg COMPONENT=topic-controller -t projectriff/topic-controller:$(TAG) -f ../Dockerfile-debug

../vendor: ../glide.lock
	cd .. && glide install -v --force

../glide.lock: ../glide.yaml
	cd .. && glide up -v --force

kubectl-apply:
	kubectl apply -f config/deployment.yaml
