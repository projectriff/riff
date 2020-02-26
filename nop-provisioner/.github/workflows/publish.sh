#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

gcloud config set disable_prompts True
gcloud auth activate-service-account --key-file <(echo ${GCLOUD_CLIENT_SECRET} | base64 --decode)
gcloud auth configure-docker

readonly root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null 2>&1 && pwd)
readonly version=$(cat ${root}/VERSION)
readonly git_branch=${GITHUB_REF:11} # drop 'refs/head/' prefix
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${GITHUB_SHA:0:16}

publishImage() {
  local tag=$1
  local source=ko.local/provisioner-735f690b778af94ebf6fae1b426bbf1d:latest
  local destination=gcr.io/projectriff/nop-provisioner/provisioner:${tag}

  docker tag ${source} ${destination}
  docker push ${destination}
}

echo "Publishing riff kafka provisioner"

publishImage ${slug}
publishImage ${version}
if [ ${git_branch} = master ] ; then
  publishImage latest
fi
