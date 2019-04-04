#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

version=`cat VERSION`
commit=$(git rev-parse HEAD)

gcloud auth activate-service-account --key-file <(echo $GCLOUD_CLIENT_SECRET | base64 --decode)

bucket=gs://projectriff/riff-cli/releases

gsutil rsync -a public-read -d ${bucket}/builds/v${version}-${commit} ${bucket}/v${version}
if [[ "$TRAVIS_BRANCH" == "master" ]]; then
  gsutil rsync -a public-read -d ${bucket}/builds/v${version}-${commit} ${bucket}/latest
fi
