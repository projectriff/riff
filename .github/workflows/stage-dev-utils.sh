#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

image=projectriff/dev-utils:${VERSION_SLUG}

GOOS=linux GOARCH=amd64 go build -o bin/publish ./cmd/publish
GOOS=linux GOARCH=amd64 go build -o bin/subscribe ./cmd/subscribe
docker build . -t ${image}

if [ $STAGE = "remote" ]; then
  docker push ${image}
fi
