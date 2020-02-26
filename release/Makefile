VERSION ?= $(shell cat VERSION)
export KAFKA_VERSION = 0.20.8

.PHONY: package
package: repository

repository: source/*.sh
	mkdir -p repository
	mkdir -p target

	./source/package.sh cert-manager ${VERSION}

	./source/package.sh contour ${VERSION}

	./source/fetch-kafka.sh kafka $(KAFKA_VERSION)
	./source/package.sh kafka ${VERSION}
	mv ./target/kafka.yaml ./target/internal-only-kafka.yaml

	./source/package.sh pulsar ${VERSION}
	mv ./target/pulsar.yaml ./target/internal-only-pulsar.yaml
	
	./source/package.sh keda ${VERSION}
	
	./source/package.sh knative ${VERSION}
	
	./source/package.sh kpack ${VERSION}
	
	./source/package.sh riff-builders ${VERSION}
	
	./source/package.sh riff-build ${VERSION}
	
	./source/package.sh riff-core-runtime ${VERSION}
	
	./source/package.sh riff-knative-runtime ${VERSION}
	
	./source/package.sh riff-streaming-runtime ${VERSION}

.PHONY: templates
templates:
	./source/update-template.sh cert-manager
	./source/update-template.sh keda
	./source/update-template.sh knative
	./source/update-template.sh kpack
	./source/update-template.sh riff-builders
	./source/update-template.sh riff-build
	./source/update-template.sh riff-core-runtime
	./source/update-template.sh riff-knative-runtime
	./source/update-template.sh riff-streaming-runtime

.PHONY: clean
clean:
	rm -rf build
	rm -rf repository
	rm -rf target
