#!/bin/bash

dir=${1}
refspec=${2:-c0db0e414c965fa36857eb993399e69db13644db} # projectriff/fats v0.2.x as of 2018-12-17

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
