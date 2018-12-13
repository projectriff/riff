#!/bin/bash

dir=${1}
refspec=${2:-257a5bab4cf4eb4153a4e96963ca5ffe9f2eaa59} # projectriff/fats master as of 2018-12-13

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
