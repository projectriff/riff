#!/bin/bash

dir=${1}
refspec=${2:-56d697a83ac1134af58f779d8c889a961a637ad5} # projectriff/fats master as of 2018-12-17

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
