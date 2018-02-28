.PHONY: build clean dockerize gen-mocks test
OUTPUT = function-controller
GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')
TAG = 0.0.5-snapshot

build: $(OUTPUT)

test:
	go test ./...

$(OUTPUT): $(GO_SOURCES)
	go build cmd/function-controller.go

gen-mocks:
	mockery -name 'TopicInformer|FunctionInformer' -dir ../kubernetes-crds/pkg/client/informers/externalversions/projectriff/v1
	mockery -name 'SharedIndexInformer' -dir ../vendor/k8s.io/client-go/tools/cache
	mockery -name 'DeploymentInformer' -dir ../vendor/k8s.io/client-go/informers/extensions/v1beta1
	mockery -name 'LagTracker|Deployer' -dir pkg/controller

clean:
	rm -f $(OUTPUT)

dockerize: $(GO_SOURCES)
	docker build .. --build-arg COMPONENT=function-controller -t projectriff/function-controller:$(TAG)

debug-dockerize: $(GO_SOURCES)
	# Need to remove probes as delve starts app in paused state
	-kubectl patch deploy/function-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/livenessProbe"}]'
	-kubectl patch deploy/function-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/readinessProbe"}]'
	docker build . -t projectriff/function-controller:$(TAG) -f Dockerfile-debug
