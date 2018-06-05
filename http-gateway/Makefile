.PHONY: build clean dockerize test kubectl-apply force-image-rebuild
OUTPUT = http-gateway
GO_SOURCES = $(shell find cmd pkg -type f -name '*.go')
IMAGE_NOT_PRESENT=$(shell docker image inspect `cat .dockerize 2>/dev/null` > /dev/null 2>&1 ; if [ $$? != 0 ] ; then echo "force-image-rebuild" ; else echo "" ; fi)
TAG ?= $(shell cat ../VERSION)

build: $(OUTPUT)

test:
	go test ./...

$(OUTPUT): $(GO_SOURCES) ../vendor
	go build cmd/http-gateway.go

gen-mocks: $(GO_SOURCES)
	go get -u github.com/vektra/mockery/.../
	go generate ./...

clean:
	rm -f $(OUTPUT) .dockerize

dockerize: .dockerize

.dockerize: $(GO_SOURCES) ../vendor ../Dockerfile ../.dockerignore $(IMAGE_NOT_PRESENT)
	docker build .. --build-arg COMPONENT=http-gateway -t projectriff/http-gateway:$(TAG) -q > .dockerize

debug-dockerize: $(GO_SOURCES) ../vendor ../Dockerfile-debug ../.dockerignore
	# Need to remove probes as delve starts app in paused state
	-kubectl patch deploy/http-gateway --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/livenessProbe"}]'
	-kubectl patch deploy/http-gateway --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/readinessProbe"}]'
	docker build .. --build-arg COMPONENT=http-gateway -t projectriff/http-gateway:$(TAG) -f ../Dockerfile-debug && rm .dockerize

../vendor: ../glide.lock
	cd .. && glide install -v --force

../glide.lock: ../glide.yaml
	cd .. && glide up -v --force

kubectl-apply:
	kubectl apply -n riff-system -f config/deployment.yaml
	kubectl apply -n riff-system -f config/service-nodeport.yaml
