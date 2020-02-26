#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly riff_version=0.6.0-snapshot

source ${FATS_DIR}/.configure.sh

${FATS_DIR}/install.sh kapp
${FATS_DIR}/install.sh ytt
${FATS_DIR}/install.sh kubectl

kubectl create ns apps

echo "Installing Cert Manager"
kapp deploy -n apps -a cert-manager -f https://storage.googleapis.com/projectriff/release/${riff_version}/cert-manager.yaml -y

source $FATS_DIR/macros/no-resource-requests.sh

echo "Installing kpack"
kapp deploy -n apps -a kpack -f https://storage.googleapis.com/projectriff/release/${riff_version}/kpack.yaml -y

echo "Installing riff Build"
kapp deploy -n apps -a riff-build -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-build.yaml -y
kapp deploy -n apps -a riff-builders -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-builders.yaml -y

echo "Installing riff Core Runtime"
kapp deploy -n apps -a riff-core-runtime -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-core-runtime.yaml -y

echo "Installing Contour"
ytt -f https://storage.googleapis.com/projectriff/release/${riff_version}/contour.yaml -f https://storage.googleapis.com/projectriff/charts/overlays/service-$(echo ${K8S_SERVICE_TYPE} | tr '[A-Z]' '[a-z]').yaml --file-mark contour.yaml:type=yaml-plain \
  | kapp deploy -n apps -a contour -f - -y

echo "Installing Knative Serving"
kapp deploy -n apps -a knative -f https://storage.googleapis.com/projectriff/release/${riff_version}/knative.yaml -y

echo "Installing riff Knative Runtime"
kapp deploy -n apps -a riff-knative-runtime -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-knative-runtime.yaml -y

echo "Installing KEDA"
kapp deploy -n apps -a keda -f https://storage.googleapis.com/projectriff/release/${riff_version}/keda.yaml -y

echo "Installing riff Streaming Runtime"
kapp deploy -n apps -a riff-streaming-runtime -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-streaming-runtime.yaml -y
