#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

mode=${1:-full}
version=`cat VERSION`
commit=$(git rev-parse HEAD)

# fetch FATS scripts
fats_dir=`dirname "${BASH_SOURCE[0]}"`/fats
fats_repo="projectriff/fats"
fats_refspec=c7c3eb7461afb68d9266e31fc43460d8a5f23da5 # projectriff/fats master as of 2019-04-08
source `dirname "${BASH_SOURCE[0]}"`/fats-fetch.sh $fats_dir $fats_refspec $fats_repo
source $fats_dir/.util.sh

$fats_dir/install.sh kubectl
$fats_dir/install.sh kail

# install riff-cli
travis_fold start install-riff
echo "Installing riff"
if [ "$mode" = "full" ]; then
  gsutil cat gs://projectriff/riff-cli/releases/builds/v${version}-${commit}/riff-linux-amd64.tgz | tar xz
  chmod +x riff
else
  make build
fi
sudo cp riff /usr/local/bin/riff
travis_fold end install-riff

# start FATS
source $fats_dir/start.sh

# install riff system
travis_fold start system-install
echo "Installing riff system"
riff system install $SYSTEM_INSTALL_FLAGS

# health checks
echo "Checking for ready pods"
wait_pod_selector_ready 'app=controller' 'knative-serving'
wait_pod_selector_ready 'app=webhook' 'knative-serving'
wait_pod_selector_ready 'app=build-controller' 'knative-build'
wait_pod_selector_ready 'app=build-webhook' 'knative-build'
echo "Checking for ready ingress"
wait_for_ingress_ready 'istio-ingressgateway' 'istio-system'

# setup namespace
kubectl create namespace $NAMESPACE
fats_create_push_credentials $NAMESPACE
riff namespace init $NAMESPACE $NAMESPACE_INIT_FLAGS
travis_fold end system-install

# run test functions
source $fats_dir/functions/helpers.sh

for test in java java-boot node npm command; do
  path=${fats_dir}/functions/uppercase/${test}
  function_name=fats-cluster-uppercase-${test}
  image=$(fats_image_repo ${function_name})
  create_args="--git-repo https://github.com/${fats_repo}.git --git-revision ${fats_refspec} --sub-path functions/uppercase/${test}"
  input_data=riff
  expected_data=RIFF

  run_function $path $function_name $image "${create_args}" $input_data $expected_data
done

for test in node command; do
  path=${fats_dir}/functions/uppercase/${test}
  function_name=fats-local-uppercase-${test}
  image=$(fats_image_repo ${function_name})
  create_args="--local-path ."
  input_data=riff
  expected_data=RIFF

  run_function $path $function_name $image "${create_args}" $input_data $expected_data
done
