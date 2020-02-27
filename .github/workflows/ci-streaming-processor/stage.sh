#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

readonly version=$(cat ${root_dir}/VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

readonly base_image="gcr.io/projectriff/streaming-processor/processor"

# stage image
./mvnw -q -B compile -Dmaven.test.skip=true
./mvnw -B com.google.cloud.tools:jib-maven-plugin:1.3.0:build \
  -Djib.to.image=${base_image}:${slug}

# stage native image
./mvnw -q -B package -Dmaven.test.skip=true
docker build -t ${base_image}-native:${slug} .
docker push ${base_image}-native:${slug}
