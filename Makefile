
.PHONY: clean build dockerize kubectl-apply

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

kubectl-apply:
	kubectl apply -f config/
	$(MAKE) -C kubernetes-crds		kubectl-apply
	$(MAKE) -C function-controller	kubectl-apply
	$(MAKE) -C http-gateway			kubectl-apply
	$(MAKE) -C topic-controller		kubectl-apply

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

