#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly version=$(cat VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

source $FATS_DIR/.configure.sh

if [ ${1:-unknown} = staged ] ; then
  echo "Using staged releases"
  release_base=https://storage.googleapis.com/projectriff/release/snapshots/${slug}
else
  echo "Using locally built releases"
  release_base=./target
fi

install_app() {
  local name=$1
  local transform=${2:-}

  if [ -z $transform ] ; then
    kapp deploy -n apps -a $name -f ${release_base}/${name}.yaml -y
  else
    ytt -f ${release_base}/${name}.yaml -f $transform --file-mark ${name}.yaml:type=yaml-plain | kapp deploy -n apps -a $name -f - -y
  fi
}

kubectl create ns apps

echo "Install Cert Manager"
install_app cert-manager

source $FATS_DIR/macros/no-resource-requests.sh

echo "Install Contour"
install_app contour .github/workflows/overlays/service-$(echo ${K8S_SERVICE_TYPE} | tr '[A-Z]' '[a-z]').yaml

echo "Install riff Build"
install_app kpack
install_app riff-builders
install_app riff-build

if [ $RUNTIME = "core" ]; then
  echo "Install riff core runtime"
  install_app riff-core-runtime
fi

if [ $RUNTIME = "knative" ]; then
  echo "Install riff Knative runtime"
  install_app knative
  install_app riff-knative-runtime
fi

if [ $RUNTIME = "streaming" ]; then
  echo "Install riff Streaming runtime"
  install_app keda
  install_app riff-streaming-runtime

  if [ $GATEWAY = "kafka" ]; then
    echo "Install Kafka"
    install_app internal-only-kafka
  fi
  if [ $GATEWAY = "pulsar" ]; then
    echo "Install Pulsar"
    install_app internal-only-pulsar
  fi
fi
