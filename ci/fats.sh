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
fats_refspec=c6dfc9a6ce1883f0db71a5d155c6a29702273265 # projectriff/fats master as of 2019-06-05
source `dirname "${BASH_SOURCE[0]}"`/fats-fetch.sh $fats_dir $fats_refspec $fats_repo
source $fats_dir/.util.sh

$fats_dir/install.sh kubectl
$fats_dir/install.sh kail
$fats_dir/install.sh duffle

# install riff-cli
travis_fold start install-riff
echo "Installing riff"
if [ "$mode" = "full" ]; then
  if [ "$machine" == "MinGw" ]; then
    curl https://storage.googleapis.com/projectriff/riff-cli/releases/builds/v${version}-${commit}/riff-windows-amd64.zip > riff.zip
    unzip riff.zip -d /usr/bin/
    rm riff.zip
  else
    curl https://storage.googleapis.com/projectriff/riff-cli/releases/builds/v${version}-${commit}/riff-linux-amd64.tgz | tar xz
    chmod +x riff
    sudo cp riff /usr/bin/riff
  fi
else
  make build
  sudo cp riff /usr/bin/riff
fi
travis_fold end install-riff

# start FATS
source $fats_dir/start.sh

# install riff system
travis_fold start system-install
echo "Installing riff system"
duffle credentials add `dirname "${BASH_SOURCE[0]}"`/duffle-creds/k8s.yaml
curl -O https://storage.googleapis.com/projectriff/riff-cnab/snapshots/riff-bundle-latest.json
duffle install riff riff-bundle-latest.json --bundle-is-file --credentials k8s --insecure

# health checks
echo "Checking for ready ingress"
wait_for_ingress_ready 'istio-ingressgateway' 'istio-system'

# setup namespace
kubectl create namespace $NAMESPACE
fats_create_push_credentials $NAMESPACE
travis_fold end system-install

# run test functions
source $fats_dir/functions/helpers.sh

if [ "$mode" = "full" ]; then
  functions=(java java-boot node npm command)
else
  functions=(command)
fi

for test in "${functions[@]}"; do
  path=${fats_dir}/functions/uppercase/${test}
  function_name=fats-cluster-uppercase-${test}
  image=$(fats_image_repo ${function_name})
  create_args="--git-repo https://github.com/${fats_repo}.git --git-revision ${fats_refspec} --sub-path functions/uppercase/${test}"
  input_data=riff
  expected_data=RIFF

  run_function $path $function_name $image "${create_args}" $input_data $expected_data
done

if [ "$machine" != "MinGw" ]; then
  for test in "${functions[@]}"; do
    path=${fats_dir}/functions/uppercase/${test}
    function_name=fats-local-uppercase-${test}
    image=$(fats_image_repo ${function_name})
    create_args="--local-path ."
    input_data=riff
    expected_data=RIFF

    run_function $path $function_name $image "${create_args}" $input_data $expected_data
  done
fi
