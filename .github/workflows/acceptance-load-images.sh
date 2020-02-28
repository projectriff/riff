#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

if [ $CLUSTER = kind ]; then
  while IFS= read -r line; do
    image=$(echo $line | awk 'BEGIN { OFS=":" } { print $1, $2 }')
    kind load docker-image $image
  done < <(docker images | grep ko.local)

  kind load docker-image projectriff/dev-utils:${VERSION_SLUG}
fi

