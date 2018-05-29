
.PHONY: build test dockerize debug-dockerize dev-setup teardown clean

build:
	$(MAKE) -C message-transport	build
	$(MAKE) -C kubernetes-crds		build
	$(MAKE) -C function-controller	build
	$(MAKE) -C function-sidecar		build
	$(MAKE) -C http-gateway			build
	$(MAKE) -C topic-controller		build
	$(MAKE) -C riff-cli				build

test:
	$(MAKE) -C message-transport	test
	$(MAKE) -C function-controller	test
	$(MAKE) -C function-sidecar		test
	$(MAKE) -C http-gateway			test
	$(MAKE) -C topic-controller		test
	$(MAKE) -C riff-cli				test

dockerize:
	$(MAKE) -C function-controller	dockerize
	$(MAKE) -C function-sidecar		dockerize
	$(MAKE) -C http-gateway			dockerize
	$(MAKE) -C topic-controller		dockerize

debug-dockerize:
	$(MAKE) -C function-controller	debug-dockerize
	$(MAKE) -C function-sidecar		debug-dockerize
	$(MAKE) -C http-gateway			debug-dockerize
	$(MAKE) -C topic-controller		debug-dockerize

dev-setup:
	kubectl apply -f config/namespace
	kubectl apply -n riff-system -f config/service-account.yaml
	kubectl apply -f config/rbac
	kubectl apply -n riff-system -f config/kafka
	$(MAKE) -C kubernetes-crds		kubectl-apply
	$(MAKE) -C function-controller	kubectl-apply
	$(MAKE) -C http-gateway			kubectl-apply
	$(MAKE) -C topic-controller		kubectl-apply

teardown:
	kubectl delete all -l function
	kubectl delete links --all
	kubectl delete topics --all
	kubectl delete functions --all
	kubectl delete all,svc -n riff-system -l app=riff
	kubectl delete crd/functions.projectriff.io
	kubectl delete crd/links.projectriff.io
	kubectl delete crd/topics.projectriff.io
	kubectl delete crd/invokers.projectriff.io
	kubectl delete -n riff-system -f config/kafka
	kubectl delete -f config/rbac
	kubectl delete -n riff-system -f config/service-account.yaml
	kubectl delete -f config/namespace

vendor: glide.lock
	glide install -v --force

glide.lock: glide.yaml
	glide up -v --force

clean:
	$(MAKE) -C function-controller	clean
	$(MAKE) -C function-sidecar		clean
	$(MAKE) -C http-gateway			clean
	$(MAKE) -C topic-controller		clean
	$(MAKE) -C riff-cli				clean
