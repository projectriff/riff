#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../.. && pwd)

test() {
    local component=$1

    echo "##[group]Stage ${component}"
    ( cd ${root_dir}/${component} && ${root_dir}/.github/workflows/test-${component}.sh )
    echo "##[endgroup]"
}

test cli
if [ $(go env GOOS) = linux ]; then
  test dev-utils
  test kafka-provisioner
  test nop-provisioner
  test pulsar-provisioner
  test stream-client-go
  test streaming-processor
  test system
fi
