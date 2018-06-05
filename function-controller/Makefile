.PHONY: build clean dockerize gen-mocks test kubectl-apply force-image-rebuild
OUTPUT = function-controller
GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')
IMAGE_NOT_PRESENT=$(shell docker image inspect `cat .dockerize 2>/dev/null` > /dev/null 2>&1 ; if [ $$? != 0 ] ; then echo "force-image-rebuild" ; else echo "" ; fi)
TAG ?= $(shell cat ../VERSION)

build: $(OUTPUT)

test: ../vendor
	go test ./...

$(OUTPUT): $(GO_SOURCES) ../vendor
	go build cmd/function-controller.go

gen-mocks: ../vendor
	go get -u github.com/vektra/mockery/.../
	mockery -name 'TopicInformer|FunctionInformer|LinkInformer' -dir ../kubernetes-crds/pkg/client/informers/externalversions/projectriff/v1alpha1
	mockery -name 'SharedIndexInformer' -dir ../vendor/k8s.io/client-go/tools/cache
	mockery -name 'DeploymentInformer' -dir ../vendor/k8s.io/client-go/informers/extensions/v1beta1
	mockery -name 'Deployer' -dir pkg/controller
	go generate ./...

# verify generated mocks which are committed or staged are up to date
verify-mocks: gen-mocks
	git diff --exit-code $(GENERATED_SOURCES)

clean:
	rm -f $(OUTPUT) .dockerize

dockerize: .dockerize

.dockerize: $(GO_SOURCES) ../vendor ../Dockerfile ../.dockerignore $(IMAGE_NOT_PRESENT)
	docker build .. --build-arg COMPONENT=function-controller -t projectriff/function-controller:$(TAG) -q > .dockerize

debug-dockerize: $(GO_SOURCES) ../vendor ../Dockerfile-debug ../.dockerignore
	# Need to remove probes as delve starts app in paused state
	-kubectl patch deploy/function-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/livenessProbe"}]'
	-kubectl patch deploy/function-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/readinessProbe"}]'
	docker build .. --build-arg COMPONENT=function-controller -t projectriff/function-controller:$(TAG) -f ../Dockerfile-debug && rm .dockerize

../vendor: ../glide.lock
	cd .. && glide install -v --force

../glide.lock: ../glide.yaml
	cd .. && glide up -v --force

kubectl-apply:
	kubectl apply -n riff-system -f config/deployment.yaml

simulation:
	go build cmd/simulator/simulator.go
	./simulator
	gnuplot -c pkg/controller/autoscaler/simulator/simulator.plt > simulation-results.png
