#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function build() {
  docker build -t ${1} .
}

function push() {
  docker tag ${1} ${2}
  docker push ${2}
}

function main() {
  ./mvnw -q -B package -Dmaven.test.skip=true

  local base_image="gcr.io/projectriff/streaming-processor/processor-native"
  local version=$(./mvnw help:evaluate -Dexpression=project.version -q -DforceStdout | tail -n1)
  local git_sha=$(git rev-parse HEAD)
  local git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
  local slug=${version}-${git_timestamp}-${git_sha:0:16}

  echo "Deploying ${base_image} (latest, ${version} and ${slug})"
  build "${base_image}"
  push "${base_image}" "${base_image}"
  push "${base_image}" "${base_image}:${version}"
  push "${base_image}" "${base_image}:${slug}"
}

main
