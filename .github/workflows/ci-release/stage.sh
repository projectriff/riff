F#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

${root_dir}/fats/install.sh helm
${root_dir}/fats/install.sh ytt
${root_dir}/fats/install.sh k8s-tag-resolver
${root_dir}/fats/install.sh yq

helm init --client-only
make clean package

if [ $STAGE = "remote" ]; then
  # upload releases
  gsutil cp -a public-read target/*.yaml gs://projectriff/release/snapshots/${VERSION_SLUG}/
fi
