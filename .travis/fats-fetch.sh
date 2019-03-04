#!/bin/bash

dir=${1}
refspec=${2:-9b43a191a7036684c174485afae35636ae8f4cfe} # projectriff/fats master as of 2019-03-04

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
