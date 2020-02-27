#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

readonly version=$(cat ${root_dir}/VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

make release

bucket=gs://projectriff/riff-cli/releases

gsutil cp -a public-read -n riff-*{.tgz,.zip} ${bucket}/builds/v${slug}/
