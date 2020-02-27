#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

readonly version=$(cat ${root_dir}/VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}


GOOS=linux GOARCH=amd64 go build -o bin/publish ./cmd/publish
GOOS=linux GOARCH=amd64 go build -o bin/subscribe ./cmd/subscribe

docker build . -t projectriff/dev-utils:${slug}
docker push projectriff/dev-utils:${slug}
