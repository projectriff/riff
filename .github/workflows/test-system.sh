#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly root_dir=$(cd `dirname $0`/../.. && pwd)
readonly fats_dir=${root_dir}/fats

${fats_dir}/install.sh kubebuilder
${fats_dir}/install.sh kustomize

make prepare test
