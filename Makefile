
.PHONY: build dockerize



build:
	$(MAKE) -C function-controller	build
	$(MAKE) -C function-sidecar		build
	$(MAKE) -C http-gateway			build
	$(MAKE) -C topic-controller		build

dockerize:
	$(MAKE) -C function-controller	dockerize
	$(MAKE) -C function-sidecar		dockerize
	$(MAKE) -C http-gateway			dockerize
	$(MAKE) -C topic-controller		dockerize


vendor: glide.lock
	glide install -v --force

glide.lock: glide.yaml
	glide up -v --force

