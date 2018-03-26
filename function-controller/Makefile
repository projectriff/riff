.PHONY: build clean dockerize gen-mocks test kubectl-apply vendor
OUTPUT = function-controller
GO_SOURCES = $(shell find pkg cmd -type f -name '*.go')
TAG ?= $(shell cat ../VERSION)

build: $(OUTPUT)

test: vendor
	go test ./...

$(OUTPUT): $(GO_SOURCES) vendor
	go build cmd/function-controller.go

gen-mocks: vendor
	mockery -name 'TopicInformer|FunctionInformer' -dir ../kubernetes-crds/pkg/client/informers/externalversions/projectriff/v1
	mockery -name 'SharedIndexInformer' -dir ../vendor/k8s.io/client-go/tools/cache
	mockery -name 'DeploymentInformer' -dir ../vendor/k8s.io/client-go/informers/extensions/v1beta1
	mockery -name 'LagTracker|Deployer' -dir pkg/controller

clean:
	rm -f $(OUTPUT)

dockerize: $(GO_SOURCES) vendor
	docker build .. --build-arg COMPONENT=function-controller -t projectriff/function-controller:$(TAG)

debug-dockerize: $(GO_SOURCES) vendor
	# Need to remove probes as delve starts app in paused state
	-kubectl patch deploy/function-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/livenessProbe"}]'
	-kubectl patch deploy/function-controller --type=json -p='[{"op":"remove", "path":"/spec/template/spec/containers/0/readinessProbe"}]'
	docker build .. --build-arg COMPONENT=function-controller -t projectriff/function-controller:$(TAG) -f ../Dockerfile-debug

vendor: 
	$(MAKE) -C .. vendor

kubectl-apply:
	kubectl apply -n riff-system  -f config/deployment.yaml
