#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

GOOS=linux GOARCH=amd64 go build -o bin/publish ./cmd/publish
GOOS=linux GOARCH=amd64 go build -o bin/subscribe ./cmd/subscribe

image=projectriff/dev-utils:${VERSION_SLUG}

docker build . -t ${image}
docker push ${image}
