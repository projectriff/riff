#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)

${root_dir}/fats/install.sh kustomize

make prepare
