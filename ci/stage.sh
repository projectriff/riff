#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

version=`cat VERSION`
commit=$(git rev-parse HEAD)

gcloud auth activate-service-account --key-file <(echo $GCLOUD_CLIENT_SECRET | base64 --decode)

make release

bucket=gs://projectriff/riff-cli/releases

gsutil cp -a public-read -n riff-*{.tgz,.zip} ${bucket}/builds/v${version}-${commit}/

