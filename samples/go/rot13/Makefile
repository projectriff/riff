
all: *.so

%.so: %.go
	docker build -f Dockerfile.plugin-builder -t plugin-builder .
	docker create --name buildenv plugin-builder:latest
	docker cp buildenv:go/$@ .
	docker rm buildenv
