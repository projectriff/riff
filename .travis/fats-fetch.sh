#!/bin/bash

dir=${1}
refspec=${2:-5b8a86b504df14acb9a03a964b8353ca4e1d8713} # projectriff/fats master as of 2018-12-20

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
