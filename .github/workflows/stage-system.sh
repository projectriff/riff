#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../.. && pwd)

${root_dir}/fats/install.sh kustomize

if [ $STAGE != "remote" ]; then
  export PROCESSOR_IMAGE_REPO=="ko.local/streaming-processor/processor"
fi

make prepare
