#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly base_image="gcr.io/projectriff/streaming-processor/processor"

# stage image
./mvnw -q -B compile -Dmaven.test.skip=true
./mvnw -B com.google.cloud.tools:jib-maven-plugin:1.3.0:build \
  -Djib.to.image=${base_image}:${VERSION_SLUG}

# stage native image
./mvnw -q -B package -Dmaven.test.skip=true
docker build -t ${base_image}-native:${VERSION_SLUG} .
docker push ${base_image}-native:${VERSION_SLUG}
