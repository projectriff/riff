#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

readonly version=$(cat ${root_dir}/VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

stageComponent() {
  local component=$1

  echo ""
  echo "# Stage riff System: ${component}"
  echo ""
  KO_DOCKER_REPO=gcr.io/projectriff/system ko resolve -P -t ${slug} -f config/riff-${component}.yaml > bin/riff-${component}.yaml
  gsutil cp -a public-read bin/riff-${component}.yaml gs://projectriff/riff-system/snapshots/riff-${component}-${slug}.yaml
}

${root_dir}/fats/install.sh ko
mkdir bin

stageComponent build
stageComponent core
stageComponent knative
stageComponent streaming
