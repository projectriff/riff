#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

readonly version=$(cat ${root_dir}/VERSION)
readonly git_branch="${1:-}"
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}


gcloud auth activate-service-account --key-file <(echo $GCLOUD_CLIENT_SECRET | base64 --decode)

bucket=gs://projectriff/riff-cli/releases

gsutil rsync -a public-read -d ${bucket}/builds/v${slug}/ ${bucket}/v${version}/
if [[ "$git_branch" == "refs/heads/master" ]]; then
  gsutil rsync -a public-read -d ${bucket}/builds/v${slug}/ ${bucket}/latest/
fi
