#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

docker run --name liiklus \
    --rm --detach \
    -p 6565:6565/tcp \
    -e storage_positions_type=MEMORY \
    -e storage_records_type=MEMORY \
    sbawaska/liiklus:20200223160346-a85402e4332c51d9

make test

docker stop liiklus
