#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

dir=${1}
refspec=${2:-master}
repo=${3:-projectriff/fats}

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/${repo}/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
