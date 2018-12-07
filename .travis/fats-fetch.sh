#!/bin/bash

dir=${1}
refspec=${2:-7eafd11ace5dc1155007f704810563d0a2a053b6} # projectriff/fats fats2 as of 2018-12-10

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
