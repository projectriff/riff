#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

docker run --name liiklus \
    --rm --detach \
    -p 6565:6565/tcp \
    -e storage_positions_type=MEMORY \
    -e storage_records_type=MEMORY \
    bsideup/liiklus:0.10.0-rc1

make test

docker stop liiklus
