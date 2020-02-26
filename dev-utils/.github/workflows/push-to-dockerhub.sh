#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly version=$(cat VERSION)
readonly git_branch=${1:11} # drop 'refs/head/' prefix
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

publishImage() {
  local tag=$1

  docker tag projectriff/dev-utils:latest projectriff/dev-utils:${tag}
  docker push projectriff/dev-utils:${tag}
}

publishImage ${slug}
publishImage ${version}
if [ $git_branch = master ] ; then
  publishImage latest
fi
