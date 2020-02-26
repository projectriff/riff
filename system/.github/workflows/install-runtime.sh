#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly version=$(cat VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

readonly riff_version=0.5.0-snapshot

source ${FATS_DIR}/.configure.sh

export KO_DOCKER_REPO=$(fats_image_repo '#' | cut -d '#' -f 1 | sed 's|/$||g')
kubectl create ns apps

echo "Installing Cert Manager"
kapp deploy -n apps -a cert-manager -f https://storage.googleapis.com/projectriff/release/${riff_version}/cert-manager.yaml -y

source $FATS_DIR/macros/no-resource-requests.sh

echo "Installing kpack"
kapp deploy -n apps -a kpack -f https://storage.googleapis.com/projectriff/release/${riff_version}/kpack.yaml -y

echo "Installing riff Build"
if [ $MODE = "push" ]; then
  kapp deploy -n apps -a riff-build -f https://storage.googleapis.com/projectriff/riff-system/snapshots/riff-build-${slug}.yaml -y
elif [ $MODE = "pull_request" ]; then
  ko resolve -f config/riff-build.yaml | kapp deploy -n apps -a riff-build -f - -y
fi
kapp deploy -n apps -a riff-builders -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-builders.yaml -y

echo "Installing Contour"
ytt -f https://storage.googleapis.com/projectriff/release/${riff_version}/contour.yaml -f https://storage.googleapis.com/projectriff/charts/overlays/service-$(echo ${K8S_SERVICE_TYPE} | tr '[A-Z]' '[a-z]').yaml --file-mark contour.yaml:type=yaml-plain \
  | kapp deploy -n apps -a contour -f - -y

if [ $RUNTIME = "core" ]; then
  echo "Installing riff Core Runtime"
  if [ $MODE = "push" ]; then
    kapp deploy -n apps -a riff-core-runtime -f https://storage.googleapis.com/projectriff/riff-system/snapshots/riff-core-${slug}.yaml -y
  elif [ $MODE = "pull_request" ]; then
    ko resolve -f config/riff-core.yaml | kapp deploy -n apps -a riff-core-runtime -f - -y
  fi
fi

if [ $RUNTIME = "knative" ]; then  
  echo "Installing Knative Serving"
  kapp deploy -n apps -a knative -f https://storage.googleapis.com/projectriff/release/${riff_version}/knative.yaml -y

  echo "Installing riff Knative Runtime"
  if [ $MODE = "push" ]; then
    kapp deploy -n apps -a riff-knative-runtime -f https://storage.googleapis.com/projectriff/riff-system/snapshots/riff-knative-${slug}.yaml -y
  elif [ $MODE = "pull_request" ]; then
    ko resolve -f config/riff-knative.yaml | kapp deploy -n apps -a riff-knative-runtime -f - -y
  fi
fi

if [ $RUNTIME = "streaming" ]; then
  echo "Installing KEDA"
  kapp deploy -n apps -a keda -f https://storage.googleapis.com/projectriff/release/${riff_version}/keda.yaml -y

  echo "Installing riff Streaming Runtime"
  if [ $MODE = "push" ]; then
    kapp deploy -n apps -a riff-streaming-runtime -f https://storage.googleapis.com/projectriff/riff-system/snapshots/riff-streaming-${slug}.yaml -y
  elif [ $MODE = "pull_request" ]; then
    ko resolve -f config/riff-streaming.yaml | kapp deploy -n apps -a riff-streaming-runtime -f - -y
  fi

  if [ $GATEWAY = "kafka" ]; then
    echo "Installing Kafka"
    kapp deploy -n apps -a internal-only-kafka -f https://storage.googleapis.com/projectriff/release/${riff_version}/internal-only-kafka.yaml -y
  fi
  if [ $GATEWAY = "pulsar" ]; then
    echo "Installing Pulsar"
    kapp deploy -n apps -a internal-only-pulsar -f https://storage.googleapis.com/projectriff/release/${riff_version}/internal-only-pulsar.yaml -y
  fi
fi
