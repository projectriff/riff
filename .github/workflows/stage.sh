#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../.. && pwd)

readonly version=$(cat VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

export VERSION_SLUG=${slug}
export STAGE=${1:-local}

stage() {
    local component=$1

    echo "##[group]Stage ${component}"
    ( cd ${root_dir}/${component} && ${root_dir}/.github/workflows/stage-${component}.sh )
    echo "##[endgroup]"
}

stage cli
stage dev-utils
stage streaming-processor
stage system
stage release
