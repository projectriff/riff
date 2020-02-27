#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

${root_dir}/fats/install.sh ko

./hack/apply-template.sh config/streaming/config/bases/processor.yaml.tpl > config/streaming/config/bases/processor.yaml
mkdir bin

stageComponent() {
  local component=$1

  echo ""
  echo "# Stage riff System: ${component}"
  echo ""
  KO_DOCKER_REPO=gcr.io/projectriff/system ko resolve -P -t ${VERSION_SLUG} -f config/riff-${component}.yaml > bin/riff-${component}.yaml
  gsutil cp -a public-read bin/riff-${component}.yaml gs://projectriff/riff-system/snapshots/riff-${component}-${VERSION_SLUG}.yaml
}

stageComponent build
stageComponent core
stageComponent knative
stageComponent streaming
