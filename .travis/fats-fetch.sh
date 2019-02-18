#!/bin/bash

dir=${1}
refspec=${2:-bd7e104fa6115147406af9c13c9fe3dce93301ed} # projectriff/fats master as of 2019-02-06

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
