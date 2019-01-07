#!/bin/bash

dir=${1}
refspec=${2:-447f0ef8d8359f6de57e392b84d00fa3e33c8f0b} # projectriff/fats master as of 2019-01-07

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
