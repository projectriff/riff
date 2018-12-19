#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

mode=${1:-full}
version=`cat VERSION`
commit=$(git rev-parse HEAD)

# fetch FATS scripts
fats_dir=`dirname "${BASH_SOURCE[0]}"`/fats
source `dirname "${BASH_SOURCE[0]}"`/fats-fetch.sh $fats_dir

# install riff-cli
if [ "$mode" = "full" ]; then
  gsutil cat gs://projectriff/riff-cli/releases/builds/v${version}-${commit}/riff-linux-amd64.tgz | tar xz
  chmod +x riff
else
  make build
fi
sudo cp riff /usr/local/bin/riff

# start FATS
source $fats_dir/start.sh

# install riff
echo "Installing riff"
riff system install $SYSTEM_INSTALL_FLAGS

# health checks
echo "Checking for ready pods"
wait_pod_selector_ready 'knative=ingressgateway' 'istio-system'
wait_pod_selector_ready 'app=controller' 'knative-serving'
wait_pod_selector_ready 'app=webhook' 'knative-serving'
wait_pod_selector_ready 'app=build-controller' 'knative-build'
wait_pod_selector_ready 'app=build-webhook' 'knative-build'
wait_pod_selector_ready 'app=eventing-controller' 'knative-eventing'
wait_pod_selector_ready 'app=webhook' 'knative-eventing'
wait_pod_selector_ready 'clusterChannelProvisioner=in-memory-channel,role=controller' 'knative-eventing'
wait_pod_selector_ready 'clusterChannelProvisioner=in-memory-channel,role=dispatcher' 'knative-eventing'
echo "Checking for ready ingress"
wait_for_ingress_ready 'knative-ingressgateway' 'istio-system'

# setup namespace
kubectl create namespace $NAMESPACE
fats_create_push_credentials $NAMESPACE
riff namespace init $NAMESPACE $NAMESPACE_INIT_FLAGS

# run test functions
echo "Run functions"
source $fats_dir/functions/helpers.sh

# uppercase
for test in java java-boot node npm command; do
  path=${fats_dir}/functions/uppercase/${test}
  function_name=fats-uppercase-${test}
  image=${IMAGE_REPOSITORY_PREFIX}/fats-uppercase-${test}:${CLUSTER_NAME}
  input_data=riff
  expected_data=RIFF

  run_function $path $function_name $image $input_data $expected_data
done

# Knative Eventing tests
source `dirname "${BASH_SOURCE[0]}"`/fats-channels.sh
