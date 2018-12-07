#!/bin/bash

dir=${1}
refspec=${2:-22f666f5b63662f92bdf4b46e175e22981ceed37} # projectriff/fats fats2 as of 2018-12-07

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
