#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if [ $STAGE = "remote" ]; then
  readonly base_image="gcr.io/projectriff/streaming-processor/processor"
  readonly jib_target=build
else
  readonly base_image="ko.local/streaming-processor/processor"
  readonly jib_target=dockerBuild
fi

# stage image
./mvnw -q -B compile -Dmaven.test.skip=true
./mvnw -B com.google.cloud.tools:jib-maven-plugin:1.3.0:dockerBuild \
  -Djib.to.image=${base_image}:${VERSION_SLUG}

# stage native image
./mvnw -q -B package -Dmaven.test.skip=true
docker build -t ${base_image}-native:${VERSION_SLUG} .
if [ $STAGE = "remote" ]; then
  docker push ${base_image}-native:${VERSION_SLUG}
fi
