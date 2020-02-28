#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

docker pull projectriff/dev-utils:${VERSION_SLUG}
docker tag projectriff/dev-utils:${VERSION_SLUG} projectriff/dev-utils:${VERSION}
docker push projectriff/dev-utils:${VERSION}
