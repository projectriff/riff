vendor: glide.lock
	glide install -v --force

glide.lock: glide.yaml
	glide up -v --force

