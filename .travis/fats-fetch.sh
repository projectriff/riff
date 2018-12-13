#!/bin/bash

dir=${1}
refspec=${2:-e07e958748330d96e8ca5f4e9a29d176374492ac} # projectriff/fats master as of 2018-12-13

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
