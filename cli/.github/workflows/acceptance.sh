#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

readonly version=$(cat ${root_dir}/VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

# fetch FATS scripts
fats_dir=`dirname "${BASH_SOURCE[0]}"`/fats
fats_repo="projectriff/fats"
fats_refspec=3d6cead12932026fdb933a1bb550cb7eca0a8def # master as of 2020-02-04
source `dirname "${BASH_SOURCE[0]}"`/fats-fetch.sh $fats_dir $fats_refspec $fats_repo
source $fats_dir/.util.sh

# install riff-cli
echo "Installing riff"
if [ "$machine" == "MinGw" ]; then
  curl https://storage.googleapis.com/projectriff/riff-cli/releases/builds/v${slug}/riff-windows-amd64.zip > riff.zip
  unzip riff.zip -d /usr/bin/
  rm riff.zip
else
  curl https://storage.googleapis.com/projectriff/riff-cli/releases/builds/v${slug}/riff-linux-amd64.tgz | tar xz
  chmod +x riff
  sudo cp riff /usr/bin/riff
fi

# start FATS
source $fats_dir/start.sh

$fats_dir/install.sh kapp
$fats_dir/install.sh ytt
kubectl create namespace apps

riff_release_version=0.5.0-snapshot

echo "Installing Cert Manager"
fats_retry kapp deploy -n apps -a cert-manager -f https://storage.googleapis.com/projectriff/release/${riff_release_version}/cert-manager.yaml -y

source $fats_dir/macros/no-resource-requests.sh

echo "Installing kpack"
kapp deploy -n apps -a kpack -f https://storage.googleapis.com/projectriff/release/${riff_release_version}/kpack.yaml -y

echo "Installing riff Build"
kapp deploy -n apps -a riff-builders -f https://storage.googleapis.com/projectriff/release/${riff_release_version}/riff-builders.yaml -y
kapp deploy -n apps -a riff-build -f https://storage.googleapis.com/projectriff/release/${riff_release_version}/riff-build.yaml -y

echo "Installing Contour"
ytt -f https://storage.googleapis.com/projectriff/release/${riff_release_version}/contour.yaml \
  -f https://storage.googleapis.com/projectriff/charts/overlays/service-$(echo ${K8S_SERVICE_TYPE} | tr '[A-Z]' '[a-z]').yaml \
  --file-mark contour.yaml:type=yaml-plain | kapp deploy -n apps -a contour -f - -y

echo "Installing Core Runtime"
kapp deploy -n apps -a riff-core-runtime -f https://storage.googleapis.com/projectriff/release/${riff_release_version}/riff-core-runtime.yaml -y

echo "Installing Knative"
kapp deploy -n apps -a knative -f https://storage.googleapis.com/projectriff/release/${riff_release_version}/knative.yaml -y

echo "Installing Knative Runtime"
kapp deploy -n apps -a riff-knative-runtime -f https://storage.googleapis.com/projectriff/release/${riff_release_version}/riff-knative-runtime.yaml -y

# setup namespace
kubectl create namespace $NAMESPACE
fats_create_push_credentials $NAMESPACE
source $fats_dir/macros/create-riff-dev-pod.sh

# run test functions
for test in command; do
  name=fats-cluster-uppercase-${test}
  image=$(fats_image_repo ${name})
  curl_opts="-H Content-Type:text/plain -H Accept:text/plain -d cli"
  expected_data="CLI"

  echo "##[group]Run function $name"

  riff function create $name --image $image --namespace $NAMESPACE --tail \
    --git-repo https://github.com/$fats_repo --git-revision $fats_refspec --sub-path functions/uppercase/${test} &

  riff core deployer create $name \
    --function-ref $name \
    --ingress-policy External \
    --namespace $NAMESPACE \
    --tail
  source $fats_dir/macros/invoke_contour.sh \
    "$(kubectl get deployers.core.projectriff.io ${name} --namespace ${NAMESPACE} -ojsonpath='{.status.url}')" \
    "${curl_opts}" \
    "${expected_data}"
  riff core deployer delete $name --namespace $NAMESPACE

  riff knative deployer create $name \
    --function-ref $name \
    --ingress-policy External \
    --namespace $NAMESPACE \
    --tail
  source $fats_dir/macros/invoke_contour.sh \
    "$(kubectl get deployers.knative.projectriff.io ${name} --namespace ${NAMESPACE} -ojsonpath='{.status.url}')" \
    "${curl_opts}" \
    "${expected_data}"
  riff knative deployer delete $name --namespace $NAMESPACE

  riff function delete $name --namespace $NAMESPACE
  fats_delete_image $image

  echo "##[endgroup]"
done

if [ "$machine" != "MinGw" ]; then
  for test in command; do
    name=fats-local-uppercase-${test}
    image=$(fats_image_repo ${name})
    curl_opts="-H Content-Type:text/plain -H Accept:text/plain -d cli"
    expected_data="CLI"

    echo "##[group]Run function $name"

    riff function create $name --image $image --namespace $NAMESPACE --tail \
      --local-path $fats_dir/functions/uppercase/${test} &

    riff core deployer create $name \
      --function-ref $name \
      --ingress-policy External \
      --namespace $NAMESPACE \
      --tail
    source $fats_dir/macros/invoke_contour.sh \
      "$(kubectl get deployers.core.projectriff.io ${name} --namespace ${NAMESPACE} -ojsonpath='{.status.url}')" \
      "${curl_opts}" \
      "${expected_data}"
    riff core deployer delete $name --namespace $NAMESPACE

    riff knative deployer create $name \
      --function-ref $name \
      --ingress-policy External \
      --namespace $NAMESPACE \
      --tail
    source $fats_dir/macros/invoke_contour.sh \
      "$(kubectl get deployers.knative.projectriff.io ${name} --namespace ${NAMESPACE} -ojsonpath='{.status.url}')" \
      "${curl_opts}" \
      "${expected_data}"
    riff knative deployer delete $name --namespace $NAMESPACE

    riff function delete $name --namespace $NAMESPACE
    fats_delete_image $image

    echo "##[endgroup]"
  done
fi
