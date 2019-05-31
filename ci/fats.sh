#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

mode=${1:-full}
version=`cat VERSION`
commit=$(git rev-parse HEAD)

# fetch FATS scripts
fats_repo="projectriff/fats"
source `dirname "${BASH_SOURCE[0]}"`/fats-fetch.sh $FATSDIR $FATSREFSPEC $fats_repo
source $FATSDIR/.util.sh

$FATSDIR/install.sh kubectl
$FATSDIR/install.sh kail
$FATSDIR/install.sh duffle

# start the cluster and registry
source $FATSDIR/start.sh

duffle init
duffle credentials add ci/myk8s.yaml
curl -O https://storage.googleapis.com/projectriff/riff-cnab/snapshots/riff-bundle-latest.json
duffle install myriff riff-bundle-latest.json --bundle-is-file --credentials myk8s --insecure

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

# create namespace
kubectl create namespace $NAMESPACE

# health checks
echo "Checking for ready ingress"
wait_for_ingress_ready 'istio-ingressgateway' 'istio-system'

# run test functions
source $FATSDIR/functions/helpers.sh

for test in java java-boot node npm command; do
  path=${FATSDIR}/functions/uppercase/${test}
  function_name=fats-cluster-uppercase-${test}
  image=$(fats_image_repo ${function_name})
  create_args="--git-repo https://github.com/${fats_repo}.git --git-revision ${FATSREFSPEC} --sub-path functions/uppercase/${test}"
  input_data=riff
  expected_data=RIFF

  run_function $path $function_name $image "${create_args}" $input_data $expected_data
done

if [ "$machine" != "MinGw" ]; then
  for test in node command; do
    path=${FATSDIR}/functions/uppercase/${test}
    function_name=fats-local-uppercase-${test}
    image=$(fats_image_repo ${function_name})
    create_args="--local-path ."
    input_data=riff
    expected_data=RIFF

    run_function $path $function_name $image "${create_args}" $input_data $expected_data
  done
fi
