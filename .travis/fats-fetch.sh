#!/bin/bash

dir=${1}
refspec=${2:-ac2638ad5df15df2a23eecd0b7ae2780d07e236b} # projectriff/fats master as of 2019-01-24

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
