#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function deploy() {
  ./mvnw -B com.google.cloud.tools:jib-maven-plugin:1.3.0:build \
    -Djib.to.image="${1}"
}

function main() {
  ./mvnw -q -B compile -Dmaven.test.skip=true

  local base_image="gcr.io/projectriff/streaming-processor/processor"
  local version=$(./mvnw help:evaluate -Dexpression=project.version -q -DforceStdout | tail -n1)
  local git_sha=$(git rev-parse HEAD)
  local git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
  local slug=${version}-${git_timestamp}-${git_sha:0:16}

  echo "Deploying ${base_image} (latest, ${version} and ${slug})"
  deploy "${base_image}"
  deploy "${base_image}:${version}"
  deploy "${base_image}:${slug}"
}

main
